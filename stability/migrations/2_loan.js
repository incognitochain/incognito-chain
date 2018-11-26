var MultiSigWallet = artifacts.require('MultiSigWallet')
var SimpleLoan = artifacts.require('SimpleLoan')

module.exports = function(deployer, network, accounts) {
    deployer.deploy(MultiSigWallet, ["0x9dC6Bb8F3Fb33FECe7a89e0b1a8655A474172152"], 1).then(() => {
        return deployer.deploy(SimpleLoan, MultiSigWallet.address)
    })
}
