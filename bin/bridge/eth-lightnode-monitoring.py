import json
import os
import time

import requests


def call_http(url, data, headers):
  resp = requests.post(url, data=json.dumps(data), headers=headers)
  resp = resp.json()
  err_key = 'error'
  if err_key in resp and resp[err_key] != None:
    print("An error occured while making http request: ", resp['error'])
    return None
  return resp

def main():
  infura_proj_id = '34918000975d4374a056ed78fe21c517'
  infura_url = 'https://mainnet.infura.io/v3/' + infura_proj_id
  header = {'Content-Type':'application/json'}

  while True:
    latest_blk_num_payload = {"jsonrpc":"2.0","method":"eth_blockNumber","params": [],"id":1}
    try:
      latest_blk_num_resp = call_http(infura_url, latest_blk_num_payload, header)
      if latest_blk_num_resp == None:
        time.sleep(600) # in sec
        continue
    except:
      print("An exception occured while calling to infura in order to get latest block number")
      continue

    try:
      latest_blk_num_dec = int(latest_blk_num_resp['result'], 16) - 15
      latest_blk_num_hex = hex(latest_blk_num_dec)

      latest_blk_payload = {"jsonrpc":"2.0","method":"eth_getBlockByNumber","params": [latest_blk_num_hex, False],"id":1}
      latest_blk_resp = call_http(infura_url, latest_blk_payload, header)
      if latest_blk_num_resp == None:
        time.sleep(600) # in sec
        continue
    except:
      print("An exception occured while calling to infura in order to get latest block number")
      continue

    try:
      blk_hash = latest_blk_resp['result']['hash']
      light_node_url = 'http://localhost:8545'
      payload = {"jsonrpc":"2.0","method":"eth_getBlockByHash","params":[blk_hash, False],"id":1}
      resp = call_http(light_node_url, payload, header)
      if resp == None:
        print('\nRestarting eth light node mainnet ')
        os.system('docker restart eth_mainnet')
      else:
        print("Everything is ok, sleeping now: ", resp)
    except:
      print('An exception raised while calling to eth light node')
      os.system('docker restart eth_mainnet')

    time.sleep(600) # in sec

main()
