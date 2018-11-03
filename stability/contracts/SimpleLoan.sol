pragma solidity ^0.4.23;

contract SimpleLoan {
    address public lender; // the only one who is able to accept collateral or return it

    enum State {Empty, Inited, Accepted, Rejected, Refunded, Liquidated}

    // TODO: each loan has separate interest and maturity window
    struct Loan {
        State state;
        address borrower;
        bytes32 digest; // must equal keccak256(key) for some bytes32 key
        uint256 amount; // in Wei
        uint256 request; // amount to loan, in CONST
        uint256 principle; // amount of loan's principle left to pay back, in CONST
        uint256 interest; // amount of interest left for current maturity cycle, in CONST
        uint256 maturityDate; // if principle is fully paid, borrower can withdraw escrow before this deadline
        uint256 escrowDeadline; // if lender doesn't accept, ETH escrow will end and collateral can be withdrawn after this deadline
        bytes stableCoinReceiver; // address to receive CONST
    }
    mapping(bytes32 => Loan) public loans;

    // TODO: update with multisig
    // TODO: support different types of loan
    uint256 public decimals = 10 ** 2; // support 2 digits after decimal point
    uint256 public loanMaturity = 90 days; // seconds
    uint256 public escrowWindow = 2 days;
    uint256 public interestRate = 1 * decimals; // 1%
    uint256 public liquidationStart = 150 * decimals; // auto-liquidation starts at 150%
    uint256 public maxPenaltyPercent = 10 * decimals; // 10%, maximum penalty for auto-liquidation

    // TODO: move to price feed contract
    uint256 public collateralPrice = 200 * decimals; // price in USD for each Wei
    uint256 public assetPrice = 1 * decimals; // price in USD for each CONST

    event __sendCollateral(bytes32 lid, uint256 amount, bytes32 offchain);
    event __acceptLoan(bytes32 lid, bytes32 key, bytes32 offchain);
    event __rejectLoan(bytes32 lid, bytes32 offchain);
    event __refundCollateral(bytes32 lid, uint256 amount, bytes32 offchain);
    event __addPayment(bytes32 lid, bytes32 offchain);
    event __liquidate(bytes32 lid, bytes32 offchain);

    constructor(address _lender) public {
        lender = _lender;
    }

    modifier onlyLender() {
        require(msg.sender == lender);
        _;
    }

    function collateralRatio(uint256 collateralAmount, uint256 debtAmount) public view returns (uint256) {
	uint256 debtValue = debtAmount * assetPrice * 10 ** 18;
	if (debtValue == 0) {
	    debtValue = 1; // dummy value in case zero debt
	}
        return collateralAmount * collateralPrice * 100 * decimals / debtValue;
    }

    function safelyCollateralized(uint256 collateralAmount, uint256 debtAmount) public view returns (bool) {
        return collateralRatio(collateralAmount, debtAmount) >= liquidationStart;
    }

    function sendCollateral(
        bytes32 lid,
        bytes32 digest,
        bytes stableCoinReceiver,
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
            require(loans[lid].state == State.Empty && safelyCollateralized(msg.value, request));

            Loan memory c;
            c.state = State.Inited;
            c.borrower = msg.sender;
            c.digest = digest;
            c.amount = msg.value;
            c.request = request;
            c.escrowDeadline = now + escrowWindow;
            c.stableCoinReceiver = stableCoinReceiver;
            loans[lid] = c;
        }

        emit __sendCollateral(lid, msg.value, offchain);
    }

    function acceptLoan(bytes32 lid, bytes32 key, bytes32 offchain) public onlyLender {
        require(loans[lid].state == State.Inited, "state must be inited");
        require(keccak256(abi.encodePacked(key)) == loans[lid].digest, "key does not match digest");
        loans[lid].state = State.Accepted;
        uint256 request = loans[lid].request;
        loans[lid].principle = request;
        loans[lid].interest = request * interestRate / 100 / decimals;
	loans[lid].maturityDate = now + loanMaturity;
        emit __acceptLoan(lid, key, offchain);
    }

    function rejectLoan(bytes32 lid, bytes32 offchain) public onlyLender {
        require(loans[lid].state == State.Inited);
        loans[lid].state = State.Rejected;
        emit __rejectLoan(lid, offchain);
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

    function addPayment(bytes32 lid, uint256 amount, bytes32 offchain) public onlyLender {
        require(loans[lid].state == State.Accepted);
        uint256 interest = loans[lid].interest;
        uint256 principle = loans[lid].principle;

        // Pay interest first if needed
	uint256 newInterest = interest;
	if (now + loanMaturity >= loans[lid].maturityDate) {
            newInterest = interest > amount ? interest - amount : 0;
            amount -= interest - newInterest; // left-over goes to principle
	}

        // Pay principle
        uint256 newPrinciple = principle > amount ? principle - amount : 0;
        loans[lid].principle = newPrinciple;

        if (newInterest != 0) {
            loans[lid].interest = newInterest;
        } else {
            loans[lid].maturityDate += loanMaturity;
            loans[lid].interest = newPrinciple * interestRate / 100 / decimals;
        }
        emit __addPayment(lid, offchain);
    }

    function liquidate(bytes32 lid, bytes32 offchain) public {
        uint256 debt = loans[lid].principle + loans[lid].interest;
        require(loans[lid].state == State.Accepted && loans[lid].principle > 0 &&
                (now > loans[lid].maturityDate || // interest wasn't paid on time
                 !safelyCollateralized(loans[lid].amount, debt))); // collateral is not enough

	uint256 currentRatio = collateralRatio(loans[lid].amount, debt);
        uint256 penaltyPercent = 0;
	if (liquidationStart > currentRatio) {
	    penaltyPercent = (liquidationStart - currentRatio) * decimals / (liquidationStart - 100 * decimals) * 5 + 5 * decimals;
	}

        if (penaltyPercent > maxPenaltyPercent) {
            penaltyPercent = maxPenaltyPercent;
        }
        uint256 liquidationPercent = penaltyPercent + 100 * decimals; // liquidate 100% debt
        uint256 liquidationAmount = debt * assetPrice * liquidationPercent / collateralPrice / 10 ** 18 / decimals / 100;
        if (liquidationAmount > loans[lid].amount) {
            liquidationAmount = loans[lid].amount; // TODO: not enough collateral?
        }

        loans[lid].principle = 0;
        loans[lid].interest = 0;
        loans[lid].amount -= liquidationAmount;
        loans[lid].state = State.Liquidated;
        lender.transfer(liquidationAmount); // TODO: transfer penalty to somewhere else?
        emit __liquidate(lid, offchain);
    }
}
