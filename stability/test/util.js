var fs = require('fs')

module.exports = {

        // convert bytes32 to string
        b2s: function(b) {
                return web3.toAscii(b).replace(/\u0000/g,'')
        },

        // convert string to bytes32
        s2b: function(s) {
                return web3.fromAscii(s)
        },

        // convert an array of string to an array of bytes32
        s2ba: function(a) {
                return a.map(s => web3.fromAscii(s))
        },

        // convert an array of string to an array of bytes32
        b2sa: function(a) {
                return a.map(b => web3.toAscii(b).replace(/\u0000/g,''))
        },

        // convert ether to wei
        e2w: function(e) {
                return web3.toWei(e)
        },

        // convert days to seconds
        d2s: function(d) {
                return d*24*60*60
        },

        // convert years to seconds
        y2s: function (y) {
                return this.d2s(y * 365);
        },

        // get onchain data
        oc: function(tx, event, key) {
                return tx.logs.filter(log => log.event == event).map(log => log.args)[0][key]
        },

        // print onchain data to the console
        poc: function(tx, event, key) {
                console.log(tx.logs.filter(log => log.event == event).map(log => log.args)[0][key])
        },

        // print onchain data to the console
        paoc: function(tx) {
                tx.logs.map(log => console.log(log.event, log.args))
        },

        // get the instance of a contract with *name* at a specific deployed *address*
        ca: function(name, address) {
                return web3.eth.contract(JSON.parse(fs.readFileSync('build/contracts/' + name + '.json').toString()).abi).at(address)
        },

        increaseTime: function(seconds) {
                web3.currentProvider.send({
                        jsonrpc: "2.0",
                        method: "evm_increaseTime",
                        params: [seconds], id: 0
                })
        },

        assertRevert: async function (promise) {
                try {
                        await promise;
                        assert.fail('Expected revert not received');
                } catch (error) {
                        const revertFound = error.message.search('revert') >= 0;
                        assert(revertFound, `Expected "revert", got ${error} instead`);
                }
        },

        eBalance: function (acc) {
                return web3.fromWei(web3.eth.getBalance(acc), 'ether').toNumber();
        },

        balance: function (acc) {
                return web3.eth.getBalance(acc).toNumber();
        },

        latestTime: function () {
                return web3.eth.getBlock('latest').timestamp;
        },

        increaseTimeTo: function (target) {
                let now = this.latestTime();
                if (target < now) throw Error(`Cannot increase current time(${now}) to a moment in the past(${target})`);
                let diff = target - now;
                this.increaseTime(diff);
        },

        gasPrice: async function (tx) {
                const txHash = tx['receipt']['transactionHash'];
                const log = await web3.eth.getTransaction(txHash);
                const gasUsed = tx['receipt']['gasUsed'];
                const gasPrice = log['gasPrice'].toNumber(); 
                return gasUsed * gasPrice;
        },

	keccak256: function(...args) {
	    args = args.map(arg => {
                if (typeof arg === 'string') {
                    if (arg.substring(0, 2) === '0x') {
                        return arg.slice(2)
                    } else {
                        return web3.toHex(arg).slice(2)
                    }
                }

                if (typeof arg === 'number') {
		    let val = (arg).toString(16)
		    while (val.length < 64) {
	                val = '0' + val
		    }
                    return val 
                } else {
                    return ''
                }
            })

            args = args.join('')
	    console.log('args:', args)
            return web3.sha3(args, { encoding: 'hex' })
        }
}
