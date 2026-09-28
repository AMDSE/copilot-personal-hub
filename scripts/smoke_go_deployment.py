import json
import time
from pathlib import Path
from http.cookies import SimpleCookie
from urllib.request import Request, urlopen
from urllib.error import HTTPError

base = "http://127.0.0.1:4143"
password = Path("/opt/m365-copilot2api/secrets/admin-password").read_text().strip()
cookie = ""

def call(path, method="GET", body=None, authenticated=False):
    headers={"Content-Type":"application/json"}
    if authenticated:
        headers["Cookie"]=cookie
    request=Request(base+path, data=json.dumps(body).encode() if body is not None else None, headers=headers, method=method)
    try:
        response=urlopen(request, timeout=25)
    except HTTPError as error:
        response=error
    return response.status,response.read().decode(),response.headers

assert call("/")[0]==200
assert call("/webapp/")[0]==200
for path in ("/api/accounts","/api/admin/keys","/v1/models"):
    assert call(path)[0]==401,path
assert call("/api/accounts/consumer","POST",{})[0]==401
status,_,headers=call("/api/admin/login","POST",{"password":password})
assert status==200,"original password login failed"
parsed=SimpleCookie()
parsed.load(headers.get("Set-Cookie",""))
assert parsed["m365_admin_session"]["httponly"]
assert parsed["m365_admin_session"]["secure"]
cookie="m365_admin_session="+parsed["m365_admin_session"].value
for path in ("/api/health","/api/version","/api/accounts","/api/admin/keys","/api/admin/settings","/api/admin/models","/api/usage","/api/m365/conversations","/api/admin/proxy-pool","/api/accounts/consumer-script"):
    assert call(path,authenticated=True)[0]==200,path
models=json.loads(call("/api/admin/models",authenticated=True)[1])["data"]
assert models and all(item["id"].startswith("copilot") for item in models)
account_id=key_id=""
try:
    snapshot={"consumer_account_id":"home:deployment-smoke."+str(int(time.time())),"username":"temporary deployment check","access_token":"fake-smoke-token-not-a-real-account","cookies":[{"name":"smoke","value":"fake-cookie","domain":"copilot.microsoft.com"}]}
    status,body,_=call("/api/accounts/consumer","POST",snapshot,True)
    assert status==200,body
    account_id=json.loads(body)["id"]
    assert call("/api/accounts/schedule","POST",{"id":account_id,"enabled":False},True)[0]==200
    accounts=call("/api/accounts",authenticated=True)[1]
    assert "fake-smoke-token" not in accounts and "fake-cookie" not in accounts
    status,body,_=call("/api/admin/keys","POST",{"name":"temporary-deployment-check"},True)
    assert status==200
    key_id=json.loads(body)["record"]["id"] if "record" in json.loads(body) else json.loads(body).get("id","")
    if not key_id:
        listing=json.loads(call("/api/admin/keys",authenticated=True)[1])["keys"]
        key_id=next(item["id"] for item in listing if item["name"]=="temporary-deployment-check")
    assert call("/api/admin/keys","PUT",{"id":key_id,"revoked":True},True)[0]==200
    print("PASS: original Go panels and admin APIs, unchanged password, Secure HttpOnly login, personal import, redacted listing, scheduling, key lifecycle, personal model catalog")
finally:
    if key_id:
        assert call("/api/admin/keys?id="+key_id,"DELETE",authenticated=True)[0]==200
    if account_id:
        assert call("/api/accounts/delete","POST",{"id":account_id},True)[0]==200
    assert call("/api/admin/logout","POST",authenticated=True)[0]==200
    print("PASS: removed only temporary smoke records; no Microsoft calls were made")
