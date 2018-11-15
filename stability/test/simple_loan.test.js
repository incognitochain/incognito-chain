const ms = artifacts.require("MultiSigWallet")
const sl = artifacts.require("SimpleLoan")

const l = console.log
const eq = assert.equal
const neq = assert.notEqual
const as = assert

const u = require('./util.js')
const fs = require('fs')
//const ww = require('web3')
var Web3 = require('web3');
var ww = new Web3(Web3.givenProvider)

contract("SimpleLoan", (accounts) => {
    const root = accounts[0]
    const owner = accounts[1]
    const requester1 = accounts[2]
    const requester2 = accounts[3]

    let c, digest, key, lid, lid1;
    let abi = null;

    before(async () => {
        s = await ms.deployed();
        c = await sl.deployed();
        key = u.padToBytes32(web3.toHex("a"))
        digest = ww.utils.soliditySha3(key) 
        abi = JSON.parse(fs.readFileSync('./build/contracts/SimpleLoan.json', 'utf8')).abi
        l(typeof(abi))
        //	l(key, digest, web3.sha3(key, { encoding: "hex" }))
    })

    let tx, tx1, tx2
    let offchain = 1

    describe('main flow', () => {
        it('gets correct collateral ratio', async () => {
            let ratio = await c.collateralRatio(web3.toWei(10), 1000, { from: requester1 })
        })

        it('should create new loan request', async () => {
            lid = 0 
            receiver = "0x123"
            request = 1000

            tx1 = await c.sendCollateral(lid, digest, receiver, request, offchain, { from: requester1, value: web3.toWei(10) })
            lid = await u.oc(tx1, "__sendCollateral", "lid")
            as(!isNaN(lid))
        })

        it('should accept loan request successfully', async () => {
            let data = c.contract.acceptLoan.getData(lid, key, offchain)
            tx1 = await s.submitTransaction(c.address, 0, data, { from: root })
            lid1 = await u.roc(tx1, abi, "__acceptLoan", "lid")
            eq(lid1, lid)
        })

        it("should add new payment", async () => {
            let amount = 100
            let eInterest = 9, ePrinciple = 910
            let data = c.contract.addPayment.getData(lid, amount, offchain)
            tx = await s.submitTransaction(c.address, 0, data, { from: root })
            lid1 = await u.roc(tx, abi, "__addPayment", "lid")
            eq(lid1, lid)
            let loan = await c.loans(lid)
            let newPrinciple = loan[5].toNumber()
            let newInterest = loan[6].toNumber()
            eq(newInterest, eInterest)
            eq(newPrinciple, ePrinciple)
        })

        it("should add another payment and wipe debt", async () => {
            let amount = 910
            let eInterest = 9, ePrinciple = 0
            let data = c.contract.addPayment.getData(lid, amount, offchain)
            tx = await s.submitTransaction(c.address, 0, data, { from: root })
            lid1 = await u.roc(tx, abi, "__addPayment", "lid")
            eq(lid1, lid)
            let loan = await c.loans(lid)
            let newPrinciple = loan[5].toNumber()
            let newInterest = loan[6].toNumber()
            eq(newInterest, eInterest)
            eq(newPrinciple, ePrinciple)
        })

        it("should be able to refund", async () => {
            tx = await c.refundCollateral(lid, offchain, { from: requester1 })
            lid1 = await u.oc(tx, "__refundCollateral", "lid")
            let amount = await u.oc(tx, "__refundCollateral", "amount")
            eq(amount, web3.toWei(10))
            eq(lid1, lid)
        })
    })

    describe('auto-liquidate after maturity date', () => {
        it('should create new loan request', async () => {
            lid = 0 
            receiver = "0x456"
            request = 1000

            tx1 = await c.sendCollateral(lid, digest, receiver, request, offchain, { from: requester1, value: web3.toWei(10) })
            lid = await u.oc(tx1, "__sendCollateral", "lid")
            as(!isNaN(lid))
        })

        it('should accept loan request successfully', async () => {
            let data = c.contract.acceptLoan.getData(lid, key, offchain)
            tx1 = await s.submitTransaction(c.address, 0, data, { from: root })
            lid1 = await u.roc(tx1, abi, "__acceptLoan", "lid")
            eq(lid1, lid)
        })

        it("should add new payment", async () => {
            let amount = 5
            let eInterest = 5, ePrinciple = 1000
            let data = c.contract.addPayment.getData(lid, amount, offchain)
            tx = await s.submitTransaction(c.address, 0, data, { from: root })
            lid1 = await u.roc(tx, abi, "__addPayment", "lid")
            eq(lid1, lid)
            let loan = await c.loans(lid)
            let newPrinciple = loan[5].toNumber()
            let newInterest = loan[6].toNumber()
            eq(newInterest, eInterest)
            eq(newPrinciple, ePrinciple)
        })

        it("should be able to liquidate", async () => {
            u.increaseTime(u.d2s(100)) // pass maturity date of loan
            tx = await c.liquidate(lid, offchain, { from: requester2 })
            let amount = await u.oc(tx, "__liquidate", "amount")
            let commission = await u.oc(tx, "__liquidate", "commission")
            eq(amount.toNumber(), web3.toWei(5.427))
            eq(commission.toNumber(), web3.toWei(0.1005))
        })

        it("should be able to refund", async () => {
            tx = await c.refundCollateral(lid, offchain, { from: requester1 })
            lid1 = await u.oc(tx, "__refundCollateral", "lid")
            let amount = await u.oc(tx, "__refundCollateral", "amount")
            eq(amount.toNumber(), web3.toWei(4.4725))
            eq(lid1, lid)
        })
    })

    describe('auto-liquidate because under-collateralized', () => {
        it('should create new loan request', async () => {
            lid = 0 
            receiver = "0x789"
            request = 1000

            tx1 = await c.sendCollateral(lid, digest, receiver, request, offchain, { from: requester1, value: web3.toWei(10) })
            lid = await u.oc(tx1, "__sendCollateral", "lid")
            as(!isNaN(lid))
        })

        it('should accept loan request successfully', async () => {
            let data = c.contract.acceptLoan.getData(lid, key, offchain)
            tx1 = await s.submitTransaction(c.address, 0, data, { from: root })
            lid1 = await u.roc(tx1, abi, "__acceptLoan", "lid")
            eq(lid1, lid)
        })
        
        it('should update price but fail to liquidate', async () => {
            let eCollateralPrice = 180 * 100
            let data = c.contract.updateCollateralPrice.getData(eCollateralPrice)
            tx1 = await s.submitTransaction(c.address, 0, data, { from: root })
            let newCollateralPrice = (await c.collateralPrice()).toNumber()
//            let loan = await c.loans(lid)
//            let amount = loan[3].toNumber(), debt = loan[5].toNumber() + loan[6].toNumber()
//            l(amount, debt)
//            l((await c.collateralRatio(amount, debt)).toNumber())
            eq(newCollateralPrice, eCollateralPrice)
            await u.assertRevert(c.liquidate(lid, offchain, { from: requester2 }))
        })

        it("should be able to liquidate", async () => {
            let eCollateralPrice = 120 * 100
            let data = c.contract.updateCollateralPrice.getData(eCollateralPrice)
            tx1 = await s.submitTransaction(c.address, 0, data, { from: root })
            tx = await c.liquidate(lid, offchain, { from: requester2 })
            let amount = await u.oc(tx, "__liquidate", "amount")
            let commission = await u.oc(tx, "__liquidate", "commission")
//            l(amount.toNumber(), commission.toNumber())
            eq(amount.toNumber(), web3.toWei(9.14252))
            as.isAtLeast(commission.toNumber(), parseInt(web3.toWei(0.115813), 10))
            as.isAtMost(commission.toNumber(), parseInt(web3.toWei(0.115814), 10))
        })
    })
})

