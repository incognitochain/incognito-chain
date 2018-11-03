var SimpleLoan = artifacts.require('SimpleLoan')

module.exports = function(deployer, network, accounts) {
        deployer.deploy(SimpleLoan, accounts[0])
}
