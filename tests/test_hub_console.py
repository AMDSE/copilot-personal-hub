from fastapi.testclient import TestClient

from m365_copilot_openai_proxy.app import create_app
from m365_copilot_openai_proxy.config import Settings


def test_console_and_assets_do_not_bypass_api_auth(tmp_path, monkeypatch):
    console = tmp_path / "console"
    console.mkdir()
    (console / "index.html").write_text("<title>Copilot Personal Hub</title>")
    assets = tmp_path / "assets"
    assets.mkdir()
    (assets / "get_token.user.js").write_text("void 0;")
    monkeypatch.setenv("HUB_CONSOLE_DIR", str(console))
    monkeypatch.setenv("HUB_ASSETS_DIR", str(assets))
    client = TestClient(create_app(Settings(TOKEN_DIR=str(tmp_path / "data"), API_KEY="test-api-key", ADMIN_PASSWORD="test-admin")))
    assert client.get("/console/").status_code == 200
    assert client.get("/hub-assets/get_token.user.js").status_code == 200
    assert client.get("/admin/accounts").status_code == 401
    assert client.get("/admin/keys").status_code == 401
    assert client.get("/v1/models").status_code == 401
    assert client.post("/admin/login", json={"password": "test-admin"}).status_code == 200
    account = client.post("/admin/accounts", json={"name": "test-personal"}).json()["account"]
    response = client.post("/admin/keys", json={"name": "test-key", "account_id": account["id"]})
    assert response.status_code == 200
    assert response.json()["key"]["account_id"] == account["id"]
    assert client.post("/admin/logout").status_code == 200
    assert client.get("/admin/keys").status_code == 401
