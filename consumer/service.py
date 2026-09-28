import asyncio
import hmac
import json
import os
import base64
from types import SimpleNamespace
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, StreamingResponse
from consumer_client import ConsumerCopilotClient, ConsumerCopilotError

app = FastAPI(docs_url=None, redoc_url=None, openapi_url=None)
turn_gate = asyncio.Semaphore(1)

@app.get("/health")
async def health():
    return {"status": "ok", "role": "consumer-transport-only"}

@app.post("/turn")
async def turn(request: Request):
    secret = os.environ.get("M365_CONSUMER_TRANSPORT_KEY", "")
    if not secret or not hmac.compare_digest(request.headers.get("authorization", ""), "Bearer " + secret):
        return JSONResponse({"error": "unauthorized"}, status_code=401)
    raw = bytearray()
    async for chunk in request.stream():
        raw.extend(chunk)
        if len(raw) > 24 * 1024 * 1024:
            return JSONResponse({"error": "request too large"}, status_code=413)
    try:
        body = json.loads(raw)
        credentials = body["credentials"]
        token = credentials["access_token"]
        cookies = {item["name"]: item["value"] for item in credentials["cookies"] if item.get("domain", "").lstrip(".") in ("", "microsoft.com", "copilot.microsoft.com")}
        prompt = str(body["prompt"])
        mode = body.get("mode", "smart")
        images = []
        for data_url in body.get("images", []):
            header, encoded = data_url.split(",", 1)
            if not header.startswith("data:image/") or ";base64" not in header:
                raise ValueError()
            if len(base64.b64decode(encoded, validate=True)) > 10 * 1024 * 1024:
                raise ValueError()
            images.append(SimpleNamespace(base64=encoded, media_type=header[5:].split(";")[0], url=""))
        if not token or mode not in ("smart", "reasoning", "chat", "search", "research", "study", "coco"):
            raise ValueError()
    except (ValueError, KeyError, TypeError):
        return JSONResponse({"error": "invalid request"}, status_code=400)
    async def events():
        try:
            async with turn_gate:
                client = ConsumerCopilotClient(cookies=cookies, access_token=token, identity_type=credentials.get("identity_type", ""), proxy=body.get("proxy") or None, mode=mode)
                async for chunk in client.chat_stream(prompt, images=images):
                    if await request.is_disconnected():
                        return
                    yield json.dumps({"delta": chunk}) + "\n"
                yield json.dumps({"done": True}) + "\n"
        except asyncio.CancelledError:
            raise
        except ConsumerCopilotError as error:
            yield json.dumps({"error": type(error).__name__ + ": upstream rejected or interrupted the request; renew credentials or check account availability"}) + "\n"
        except Exception:
            yield json.dumps({"error": "personal transport failed; credentials are never included in errors"}) + "\n"
    return StreamingResponse(events(), media_type="application/x-ndjson", headers={"Cache-Control": "no-store"})
