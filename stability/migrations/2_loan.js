var MultiSigWallet = artifacts.require('MultiSigWallet')
var SimpleLoan = artifacts.require('SimpleLoan')

module.exports = function(deployer, network, accounts) {
    deployer.deploy(MultiSigWallet, [accounts[0]], 1).then(() => {
        return deployer.deploy(SimpleLoan, MultiSigWallet.address, accounts[1])
    })
}
