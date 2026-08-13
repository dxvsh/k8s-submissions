import json
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse

LOG_FILE = "logs/logs.txt"
HOST = "0.0.0.0"
PORT = int(os.environ.get("PORT", "9090"))


def read_latest_log_entry():
    latest_line = None

    with open(LOG_FILE, "r", encoding="utf-8") as log_file:
        for line in log_file:
            stripped_line = line.strip()
            if stripped_line:
                latest_line = stripped_line

    if latest_line is None:
        return None

    return json.loads(latest_line)


class StatusHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if urlparse(self.path).path != "/status":
            self.send_error(404)
            return

        try:
            response = read_latest_log_entry()
        except FileNotFoundError:
            self.send_json_response(
                503,
                {"error": "logs.txt does not exist yet"},
            )
            return
        except json.JSONDecodeError:
            self.send_json_response(
                500,
                {"error": "latest log line is not valid JSON"},
            )
            return

        if response is None:
            self.send_json_response(
                503,
                {"error": "logs.txt does not contain any log entries yet"},
            )
            return

        self.send_json_response(200, response)

    def send_json_response(self, status_code, response):
        body = json.dumps(response).encode("utf-8")

        self.send_response(status_code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


if __name__ == "__main__":
    server = ThreadingHTTPServer((HOST, PORT), StatusHandler)
    server.serve_forever()
