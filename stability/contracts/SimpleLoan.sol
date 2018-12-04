pragma solidity ^0.5.0;

contract SimpleLoan {
    address payable public lender; // the one who is able to accept collateral and return it
    address payable public owner; // creator of the contract, able to lock collateral and update payment on behalf of lender

    enum State {Empty, Inited, Accepted, Rejected, Refunded, Liquidated}

    // TODO: each loan has separate interest and maturity window
    struct Loan {
        State state;
        address payable borrower;
        bytes32 digest; // must equal keccak256(key) for some bytes32 key
        uint256 amount; // in Wei
        uint256 request; // amount to loan, in CONST
        uint256 principle; // amount of loan's principle left to pay back, in CONST
        uint256 escrowDeadline; // if lender doesn't accept, ETH escrow will end and collateral can be withdrawn after this deadline
        bytes stableCoinReceiver; // address to receive CONST
    }
    mapping(bytes32 => Loan) public loans;

    mapping(bytes32 => uint256) public params;
    uint256 public decimals = 10 ** 2; // support 2 digits after decimal point

    event __sendCollateral(bytes32 lid, uint256 amount, bytes32 offchain);
    event __acceptLoan(bytes32 lid, bytes32 key, bytes32 offchain);
    event __rejectLoan(bytes32 lid, bytes32 offchain);
    event __refundCollateral(bytes32 lid, uint256 amount, bytes32 offchain);
    event __liquidate(bytes32 lid, uint256 amount, bytes32 offchain);
    event __wipeDebt(bytes32 lid, bytes32 offchain);
    event __update(bytes32 name, uint256 value, bytes32 offchain);

    modifier onlyLender() {
        require(msg.sender == lender, "only lenders are authorized");
        _;
    }

    modifier lenderOrOwner() {
        require(msg.sender == lender || msg.sender == owner, "only lenders and owner are authorized");
        _;
    }

    constructor(address payable _lender, address payable _owner) public {
        lender = _lender;
        owner = _owner;

        // TODO: support different types of loan
        params["loanMaturity"] = 90 days; // seconds
        params["escrowWindow"] = 3 days;
        params["interestRate"] = 1 * decimals; // 1%
        params["liquidationStart"] = 150 * decimals; // auto-liquidation starts at 150%
        params["liquidationEnd"] = 100 * decimals; // below 100%, commission doesn't increase
        params["liquidationPenalty"] = 10 * decimals; // 10%, maximum penalty for auto-liquidation
    }

    function update(bytes32 name, uint256 value, bytes32 offchain) public onlyLender {
        params[name] = value;
        emit __update(name, value, offchain);
    }

    function get(bytes32 name) public view returns (uint256) {
        return params[name];
    }

    function part(uint256 value, uint256 percent) public view returns (uint256) {
        return value * percent / decimals / 100;
    }

    function collateralRatio(uint256 collateralAmount, uint256 debtAmount, uint256 collateralPrice, uint256 assetPrice) public view returns (uint256) {
        uint256 debtValue = debtAmount * assetPrice * 10 ** 18;
        if (debtValue == 0) {
            debtValue = 1; // dummy value in case zero debt
        }
        return collateralAmount * collateralPrice * 100 * decimals / debtValue;
    }

    function safelyCollateralized(uint256 collateralAmount, uint256 debtAmount, uint256 collateralPrice, uint256 assetPrice) public view returns (bool) {
        return collateralRatio(collateralAmount, debtAmount, collateralPrice, assetPrice) >= params["liquidationStart"];
    }

    function sendCollateral(
        bytes32 lid,
        bytes32 digest,
        bytes memory stableCoinReceiver,
        uint256 request,
        bytes32 offchain
    )
    public
    payable
    {
        if (lid != 0x0) {
            // TODO: allow update request amount of CONST?
            // TODO: update escrowDeadline?
            require(loans[lid].state == State.Inited || loans[lid].state == State.Accepted);
            loans[lid].amount += msg.value;
        } else {
            lid = keccak256(abi.encodePacked(digest, stableCoinReceiver, request));
            require(loans[lid].state == State.Empty);

            Loan memory c;
            c.state = State.Inited;
            c.borrower = msg.sender;
            c.digest = digest;
            c.amount = msg.value;
            c.request = request;
            c.escrowDeadline = now + params["escrowWindow"];
            c.stableCoinReceiver = stableCoinReceiver;
            loans[lid] = c;
        }

        emit __sendCollateral(lid, msg.value, offchain);
    }

    function acceptLoan(bytes32 lid, bytes32 key, bytes32 offchain) public lenderOrOwner {
        // TODO: loan hasn't passed escrow deadline?
        require(loans[lid].state == State.Inited, "state must be inited");
        require(keccak256(abi.encodePacked(key)) == loans[lid].digest, "key does not match digest");
        loans[lid].state = State.Accepted;
        uint256 request = loans[lid].request;
        loans[lid].principle = request;
        emit __acceptLoan(lid, key, offchain);
    }

    function rejectLoan(bytes32 lid, bytes32 offchain) public lenderOrOwner {
        require(loans[lid].state == State.Inited);
        loans[lid].state = State.Rejected;
        emit __rejectLoan(lid, offchain);
        refundCollateral(lid, offchain);
    }

    function refundCollateral(bytes32 lid, bytes32 offchain) public {
        require(loans[lid].state == State.Rejected ||
                loans[lid].state == State.Liquidated ||
                (loans[lid].state == State.Inited && now > loans[lid].escrowDeadline) ||
                (loans[lid].state == State.Accepted && loans[lid].principle == 0));
        loans[lid].state = State.Refunded;
        uint256 amount = loans[lid].amount;
        loans[lid].amount = 0;
        loans[lid].borrower.transfer(amount);
        emit __refundCollateral(lid, amount, offchain);
    }

    function wipeDebt(bytes32 lid, bytes32 offchain) public lenderOrOwner {
        require(loans[lid].state == State.Accepted);
        loans[lid].principle = 0;
        emit __wipeDebt(lid, offchain);
    }

    function liquidate(bytes32 lid, uint256 interest, uint256 collateralPrice, uint256 assetPrice, bytes32 offchain) public lenderOrOwner {
        uint256 debt = loans[lid].principle + interest;
        require(loans[lid].state == State.Accepted && loans[lid].principle > 0 &&
                !safelyCollateralized(loans[lid].amount, debt, collateralPrice, assetPrice)); // collateral is not enough

        uint256 base = debt * assetPrice * 10 ** 18 / collateralPrice; // Wei amount needed to buy back enough Constant at current price
        uint256 penalty = part(base, params["liquidationPenalty"]);
        uint256 liquidationAmount = base + penalty;
        uint256 collateralAmount = loans[lid].amount;
        if (liquidationAmount > collateralAmount) {
            liquidationAmount = collateralAmount; // TODO: not enough collateral?
        }
        loans[lid].principle = 0;
        loans[lid].amount -= liquidationAmount;
        loans[lid].state = State.Liquidated;
        owner.transfer(liquidationAmount); // TODO: transfer penalty to somewhere else?
        emit __liquidate(lid, liquidationAmount, offchain);
    }
}
