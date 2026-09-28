import json
from fastapi.testclient import TestClient
import service


def test_private_transport_rejects_missing_secret(monkeypatch):
    monkeypatch.setenv("M365_CONSUMER_TRANSPORT_KEY", "test-private-secret")
    client = TestClient(service.app)
    assert client.get("/health").status_code == 200
    assert client.post("/turn", json={}).status_code == 401


def test_transport_stream_and_no_account_storage(monkeypatch):
    monkeypatch.setenv("M365_CONSUMER_TRANSPORT_KEY", "test-private-secret")
    class FakeConsumer:
        def __init__(self, **kwargs):
            assert kwargs["access_token"] == "dummy-token"
            assert kwargs["mode"] == "reasoning"
        async def chat_stream(self, prompt, images):
            assert prompt == "history preserved"
            yield "first "
            yield "second"
    monkeypatch.setattr(service, "ConsumerCopilotClient", FakeConsumer)
    client = TestClient(service.app)
    response = client.post("/turn", headers={"Authorization": "Bearer test-private-secret"}, json={"credentials": {"access_token": "dummy-token", "cookies": [{"name": "session", "value": "secret", "domain": ".copilot.microsoft.com"}]}, "prompt": "history preserved", "mode": "reasoning"})
    assert response.status_code == 200
    assert [json.loads(line) for line in response.text.splitlines()] == [{"delta": "first "}, {"delta": "second"}, {"done": True}]


def test_transport_redacts_upstream_exception(monkeypatch):
    monkeypatch.setenv("M365_CONSUMER_TRANSPORT_KEY", "test-private-secret")
    class FailedConsumer:
        def __init__(self, **kwargs):
            pass
        async def chat_stream(self, prompt, images):
            raise service.ConsumerCopilotError("do-not-leak-bearer-token")
            yield ""
    monkeypatch.setattr(service, "ConsumerCopilotClient", FailedConsumer)
    response = TestClient(service.app).post("/turn", headers={"Authorization": "Bearer test-private-secret"}, json={"credentials": {"access_token": "dummy-token", "cookies": []}, "prompt": "test"})
    assert "do-not-leak-bearer-token" not in response.text
    assert "error" in json.loads(response.text)
