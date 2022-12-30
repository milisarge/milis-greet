#!/usr/bin/env python3
import sys
import socket, os
import json

soket = os.getenv("GREETD_SOCK")
client = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)

client.connect(soket)

def g_send(json_req):
  req = json.dumps(jreq) 
  client.send( len(req).to_bytes(4,"little")  + req.encode())
  return client.recv(128)

start = 1
while True:
  try:
    if start == 1:
      username = input("user:")
      jreq = {"type": "create_session", "username": username }
      g_send(jreq)
      start = 2
    if start == 2:
      password = input("password:")
      jreq = {"type": "post_auth_message_response", "response": password}
      resp = g_send(jreq)
      resp = json.loads(resp.decode("utf-8"))
      start = 3
    if start == 3:
      cmd = input("cmd:")
      jreq = {"type": "start_session", "cmd": cmd.split() }
      resp_raw = g_send(jreq)
      resp_len = int.from_bytes(resp_raw[0:4],"little")
      respt = resp_raw[4:resp_len+4].decode()
      resp = json.loads(respt)
      if "error_type" in resp and resp["error_type"] == "auth_error":
        start = 1
        print("auth error - try again")      
      elif "type" in resp and resp["type"] == "success":
        print(resp)
        sys.exit()
  except KeyboardInterrupt as k:
    client.close()
    break
