const ms = artifacts.require("MultiSigWallet")
const rs = artifacts.require("Reserve")

const l = console.log
const eq = assert.equal
const neq = assert.notEqual
const as = assert

const u = require('./util.js')
const fs = require('fs')
//const ww = require('web3')
var Web3 = require('web3');
var ww = new Web3(Web3.givenProvider)

contract("Reserve", (accounts) => {
    const msAcc = accounts[0]
    const owner = accounts[1]
    const requester1 = accounts[2]
    const requester2 = accounts[3]

    let c, digest, key, lid, lid1;
    let abi = null;
    let tx, tx1, tx2
    let offchain = "0x456"

    before(async () => {
        s = await ms.deployed();
        c = await rs.deployed();
        abi = JSON.parse(fs.readFileSync('./build/contracts/Reserve.json', 'utf8')).abi
    })

    function getFunc(abiObj, name) {
        for (var i = 0; i < abiObj.length; ++i) {
            if (abi[i].name == name) {
                return abi[i]
            }
        }
    }

    describe('main flow', () => {
        it('should raise reserve successfully', async () => {
            receiver = "0x123"

            tx1 = await c.raise(receiver, offchain, { from: requester1, value: web3.utils.toWei("10") })
            value = await u.oc(tx1, "__raise", "value")
            eq(value.toString(), web3.utils.toWei("10"))
        })

        it('should fail to spend', async () => {
            await u.assertRevert(c.spend(web3.utils.toWei("1"), offchain, { from: requester1 } ))
        })

        it('should spend reserve successfully', async () => {
            tx1 = await c.spend(web3.utils.toWei("1"), offchain, { from: owner } )
            let amount = await u.oc(tx1, "__spend", "amount")
            eq(amount.toString(), web3.utils.toWei("1"))
        })
    })
})

