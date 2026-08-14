import json
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.error import HTTPError, URLError
from urllib.parse import urlparse
from urllib.request import urlopen

LOG_FILE = "logs/logs.txt"
PINGPONG_SVC_NAME = os.environ.get("PINGPONG_SVC_NAME", "pingpong-svc")
PINGPONG_SVC_PORT = os.environ.get("PINGPONG_SVC_PORT", "4567")
PINGS_URL = f"http://{PINGPONG_SVC_NAME}:{PINGPONG_SVC_PORT}/pings"
PINGPONG_REQUEST_TIMEOUT_SECONDS = 5
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


def read_ping_count():
    with urlopen(PINGS_URL, timeout=PINGPONG_REQUEST_TIMEOUT_SECONDS) as response:
        ping_response = json.loads(response.read().decode("utf-8"))
        return int(ping_response["ping_count"])


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

        try:
            response["Ping / Pongs"] = read_ping_count()
        except HTTPError as error:
            self.send_json_response(
                502,
                {"error": f"ping service returned HTTP {error.code}"},
            )
            return
        except URLError:
            self.send_json_response(
                502,
                {"error": "could not reach ping service"},
            )
            return
        except (KeyError, json.JSONDecodeError, ValueError):
            self.send_json_response(
                500,
                {"error": "ping service did not return a valid ping_count"},
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
