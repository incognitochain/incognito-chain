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
	key = "" 
	digest = ww.utils.soliditySha3(key) 
    })

    let tx1, hid1
    let tx2, hid2
    let offchain = 1

    describe('main flow', () => {
        it('gets correct collateral ratio', async () => {
            let ratio = await c.collateralRatio(web3.toWei(10), 1000, { from: requester1 })
        })
	    
        it('should create new c request', async () => {
	    lid = 0 
	    receiver = 0 
	    request = 1000

            tx1 = await c.sendCollateral(lid, digest, receiver, request, offchain, { from: requester1, value: web3.toWei(10) })
            lid = await u.oc(tx1, "__sendCollateral", "lid")
            as(!isNaN(lid))
        })

        it('should accept c request successfully', async () => {
            tx1 = await c.acceptLoan(lid, key, offchain, { from: root })
            lid1 = await u.oc(tx1, "__acceptLoan", "lid")
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

//    describe('when inited', () => {
//        it('should be able to make handshake from second entity', async () => {
//            tx1 = await c.init(owner2, offchain, { from: owner1 })
//            hid1 = await oc(tx1, "__init", "hid")
//            tx2 = await c.shake(hid1, offchain, { from: owner2 })
//            hid2 = await oc(tx2, "__shake", "hid")
//            eq(Number(hid1), Number(hid2))
//        })
//
//        it("should fail to shake if acceptor does not match", async () => {
//            tx1 = await c.init(owner3, offchain, { from: owner2 })
//            hid1 = await oc(tx1, "__init", "hid")
//            u.assertRevert(c.shake(hid1, offchain, { from: owner1 } ))
//        })
//
//        it('should update acceptor if not set when init', async () => {
//            tx1 = await c.init('0x0', offchain, { from: owner1 })
//            hid1 = await oc(tx1, "__init", "hid")
//            eq((await c.handshakes(hid1))[1], 0)
//
//            tx2 = await c.shake(hid1, offchain, { from: owner2 })
//            await oc(tx2, "__shake", "hid")
//            eq((await c.handshakes(hid1))[1], owner2)
//        })
//    })
})

