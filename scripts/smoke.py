import argparse
import json
import time
from http.cookies import SimpleCookie
from pathlib import Path
from urllib.error import HTTPError
from urllib.request import Request, urlopen


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--url", default="http://127.0.0.1:4142")
    parser.add_argument("--password-file", required=True)
    options = parser.parse_args()
    password = Path(options.password_file).read_text().strip()
    cookie = ""

    def call(path, method="GET", body=None, authenticated=False):
        headers = {"Content-Type": "application/json"}
        if authenticated:
            headers["Cookie"] = cookie
        request = Request(options.url.rstrip("/") + path, data=json.dumps(body).encode() if body is not None else None, headers=headers, method=method)
        try:
            response = urlopen(request, timeout=30)
        except HTTPError as error:
            response = error
        payload = response.read().decode()
        return response.status, payload, response.headers

    for path in ("/healthz", "/console/", "/hub-assets/get_token.user.js"):
        assert call(path)[0] == 200, path
    for path in ("/admin/accounts", "/admin/keys", "/v1/models"):
        assert call(path)[0] == 401, path
    status, _, headers = call("/admin/login", "POST", {"password": password})
    assert status == 200, "admin login failed"
    parsed = SimpleCookie()
    parsed.load(headers.get("Set-Cookie", ""))
    assert parsed["admin_auth"]["httponly"]
    assert parsed["admin_auth"]["secure"]
    cookie = "admin_auth=" + parsed["admin_auth"].value
    account_id = key_id = ""
    try:
        status, body, _ = call("/admin/accounts", "POST", {"name": "smoke-" + str(int(time.time()))}, True)
        assert status == 200
        account_id = json.loads(body)["account"]["id"]
        status, body, _ = call("/admin/keys", "POST", {"name": "smoke-key", "account_id": account_id}, True)
        assert status == 200
        key_id = json.loads(body)["key"]["id"]
        assert call("/admin/keys/" + key_id, "POST", {"enabled": False}, True)[0] == 200
        assert call("/admin/accounts", authenticated=True)[0] == 200
        print("PASS: health, console, script, anonymous API denial, secure admin login, account/key creation and disable")
    finally:
        if key_id:
            assert call("/admin/keys/" + key_id, "DELETE", authenticated=True)[0] == 200
        if account_id:
            assert call("/admin/accounts/" + account_id, "DELETE", authenticated=True)[0] == 200
        assert call("/admin/logout", "POST")[0] == 200
        print("PASS: temporary smoke account/key removed; Microsoft chat remains untested until owner authorization")


if __name__ == "__main__":
    main()
