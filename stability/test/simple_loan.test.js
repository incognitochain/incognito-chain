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
    const msAcc = accounts[0]
    const owner = accounts[1]
    const requester1 = accounts[2]
    const requester2 = accounts[3]

    let c, digest, key, lid, lid1;
    let abi = null;

    before(async () => {
        s = await ms.deployed();
        c = await sl.deployed();
        key = u.padToBytes32(web3.utils.toHex("a"))
        digest = ww.utils.soliditySha3(key) 
        abi = JSON.parse(fs.readFileSync('./build/contracts/SimpleLoan.json', 'utf8')).abi
        //	l(key, digest, web3.sha3(key, { encoding: "hex" }))
    })

    function getFunc(abiObj, name) {
        for (var i = 0; i < abiObj.length; ++i) {
            if (abi[i].name == name) {
                return abi[i]
            }
        }
    }

    let tx, tx1, tx2
    let offchain = "0x456"

    describe('main flow', () => {
        it('should create new loan request', async () => {
            lid = "0x0"
            receiver = "0x123"
            request = 1000

            tx1 = await c.sendCollateral(lid, digest, receiver, request, offchain, { from: requester1, value: web3.utils.toWei("10") })
            lid = await u.oc(tx1, "__sendCollateral", "lid")
            as(!isNaN(lid))
        })

        it('should accept loan request successfully', async () => {
            let data = web3.eth.abi.encodeFunctionCall(getFunc(abi, "acceptLoan"), [lid, key, offchain])
            tx1 = await s.submitTransaction(c.address, 0, data, { from: msAcc })
            lid1 = await u.roc(tx1, abi, "__acceptLoan", "lid")
            eq(lid1, lid)
        })

        it("should add new payment", async () => {
            let amount = 100
            let eInterest = 9, ePrinciple = 910
            let data = web3.eth.abi.encodeFunctionCall(getFunc(abi, "addPayment"), [lid, amount, offchain])
            tx = await s.submitTransaction(c.address, 0, data, { from: msAcc })
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
            let data = web3.eth.abi.encodeFunctionCall(getFunc(abi, "addPayment"), [lid, amount, offchain])
            tx = await s.submitTransaction(c.address, 0, data, { from: msAcc })
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
            eq(amount, web3.utils.toWei("10"))
            eq(lid1, lid)
        })
    })

    describe('auto-liquidate after maturity date', () => {
        it('should create new loan request', async () => {
            lid = "0x0"
            receiver = "0x789"
            request = 1000

            tx1 = await c.sendCollateral(lid, digest, receiver, request, offchain, { from: requester1, value: web3.utils.toWei("10") })
            lid = await u.oc(tx1, "__sendCollateral", "lid")
            as(!isNaN(lid))
        })

        it('should accept loan request successfully', async () => {
            let data = web3.eth.abi.encodeFunctionCall(getFunc(abi, "acceptLoan"), [lid, key, offchain])
            tx1 = await s.submitTransaction(c.address, 0, data, { from: msAcc })
            lid1 = await u.roc(tx1, abi, "__acceptLoan", "lid")
            eq(lid1, lid)
        })

        it("should add new payment", async () => {
            let amount = 5
            let eInterest = 5, ePrinciple = 1000
            let data = web3.eth.abi.encodeFunctionCall(getFunc(abi, "addPayment"), [lid, amount, offchain])
            tx = await s.submitTransaction(c.address, 0, data, { from: msAcc })
            lid1 = await u.roc(tx, abi, "__addPayment", "lid")
            eq(lid1, lid)
            let loan = await c.loans(lid)
            let newPrinciple = loan[5].toNumber()
            let newInterest = loan[6].toNumber()
            eq(newInterest, eInterest)
            eq(newPrinciple, ePrinciple)
        })

        it("should fail to liquidate", async () => {
            u.increaseTime(u.d2s(100)) // pass maturity date of loan
            await u.assertRevert(c.liquidate(lid, 12000, 100, offchain, { from: requester2 })) // Caller not authorized
        })

        it("should be able to liquidate", async () => {
            let x = await c.loans(lid)
            tx = await c.liquidate(lid, 12000, 100, offchain, { from: owner })
            x = await c.loans(lid)
            let amount = await u.oc(tx, "__liquidate", "amount")
            eq(amount.toString(), web3.utils.toWei("9.2125"))
        })

        it("should be able to refund", async () => {
            tx = await c.refundCollateral(lid, offchain, { from: requester1 })
            lid1 = await u.oc(tx, "__refundCollateral", "lid")
            let amount = await u.oc(tx, "__refundCollateral", "amount")
            eq(amount.toString(), web3.utils.toWei("0.7875"))
            eq(lid1, lid)
        })
    })

    describe('auto-liquidate because under-collateralized', () => {
        it('should create new loan request', async () => {
            lid = "0x0"
            receiver = "0xabc"
            request = 1000

            tx1 = await c.sendCollateral(lid, digest, receiver, request, offchain, { from: requester1, value: web3.utils.toWei("10") })
            lid = await u.oc(tx1, "__sendCollateral", "lid")
            as(!isNaN(lid))
        })

        it('should let owner accept loan request successfully', async () => {
            tx1 = await c.acceptLoan(lid, key, offchain, { from: owner })
            lid1 = await u.roc(tx1, abi, "__acceptLoan", "lid")
            eq(lid1, lid)
        })
        
        it('should update price but fail to liquidate', async () => {
            let collateralPrice = 180 * 100
            let assetPrice = 1 * 100
            await u.assertRevert(c.liquidate(lid, collateralPrice, assetPrice, offchain, { from: msAcc }))
        })

        it("should be able to liquidate", async () => {
            let collateralPrice = 120 * 100
            let assetPrice = 1 * 100
            let data = web3.eth.abi.encodeFunctionCall(getFunc(abi, "liquidate"), [lid, collateralPrice, assetPrice, offchain])
            tx = await s.submitTransaction(c.address, 0, data, { from: msAcc })
            let amount = await u.roc(tx, abi, "__liquidate", "amount")
            eq(amount.toString().substr(0, 5), web3.utils.toWei("9.2583333").substr(0, 5))
        })
    })
})

