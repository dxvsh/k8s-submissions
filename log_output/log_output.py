import json
import os
import uuid
from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

UUID_STRING = str(uuid.uuid4())
HOST = "0.0.0.0"
PORT = int(os.environ.get("PORT", "9090"))


class StatusHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path != "/status":
            self.send_error(404)
            return

        response = {
            "uuid": UUID_STRING,
            "timestamp": datetime.now(timezone.utc).isoformat(),
        }
        body = json.dumps(response).encode("utf-8")

        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


if __name__ == "__main__":
    server = ThreadingHTTPServer((HOST, PORT), StatusHandler)
    server.serve_forever()
