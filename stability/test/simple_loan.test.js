const p2p = artifacts.require("SimpleLoan")

const l = console.log
const eq = assert.equal
const neq = assert.notEqual
const as = assert

const u = require('./util.js')
const ww = require('web3')

contract("SimpleLoan", (accounts) => {
    const root = accounts[0]
    const owner = accounts[1]
    const requester1 = accounts[2]

    let c, digest, key, lid;

    before(async () => {
        c = await p2p.deployed();
        key = u.padToBytes32(web3.toHex("a"))
        digest = ww.utils.soliditySha3(key) 
        //	l(key, digest, web3.sha3(key, { encoding: "hex" }))
    })

    let tx1, tx2
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
            l(lid)
            as(!isNaN(lid))
        })

        it('should accept loan request successfully', async () => {
            tx1 = await c.acceptLoan(lid, key, offchain, { from: root })
            lid1 = await u.oc(tx1, "__acceptLoan", "lid")
//            l((await u.oc(tx0, "__acceptLoan", "offchain")))
//            l((await u.oc(tx1, "__acceptLoan", "key")))
            eq(lid1, lid)
        })

        it("should add new payment", async () => {
            let amount = 100
            let eInterest = 9, ePrinciple = 910
            tx = await c.addPayment(lid, amount, offchain, { from: root })
            lid1 = await u.oc(tx, "__addPayment", "lid")
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
            tx = await c.addPayment(lid, amount, offchain, { from: root })
            lid1 = await u.oc(tx, "__addPayment", "lid")
            eq(lid1, lid)
            let loan = await c.loans(lid)
            let newPrinciple = loan[5].toNumber()
            let newInterest = loan[6].toNumber()
            eq(newInterest, eInterest)
            eq(newPrinciple, ePrinciple)
        })

        it("should be able to refund", async () => {
            tx = await c.refundCollateral(lid, offchain, { from: root })
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
            l(lid)
            as(!isNaN(lid))
        })

        it('should accept loan request successfully', async () => {
            tx1 = await c.acceptLoan(lid, key, offchain, { from: root })
            lid1 = await u.oc(tx1, "__acceptLoan", "lid")
            eq(lid1, lid)
        })

        it("should add new payment", async () => {
            let amount = 5
            let eInterest = 5, ePrinciple = 1000
            tx = await c.addPayment(lid, amount, offchain, { from: root })
            lid1 = await u.oc(tx, "__addPayment", "lid")
            eq(lid1, lid)
            let loan = await c.loans(lid)
            let newPrinciple = loan[5].toNumber()
            let newInterest = loan[6].toNumber()
            eq(newInterest, eInterest)
            eq(newPrinciple, ePrinciple)
        })

        it("should be able to liquidate", async () => {
            u.increaseTime(u.d2s(100)) // pass maturity date of loan
            tx = await c.liquidate(lid, offchain, { from: root })
            lid1 = await u.oc(tx, "__liquidate", "lid")
            let loan = await c.loans(lid)
            let amount = loan[3].toNumber()
            eq(amount, web3.toWei(10))
            eq(lid1, lid)
        })

        it("should be able to refund", async () => {
            tx = await c.refundCollateral(lid, offchain, { from: root })
            lid1 = await u.oc(tx, "__refundCollateral", "lid")
            let amount = await u.oc(tx, "__refundCollateral", "amount")
            eq(amount, web3.toWei(10))
            eq(lid1, lid)
        })
    })
})

