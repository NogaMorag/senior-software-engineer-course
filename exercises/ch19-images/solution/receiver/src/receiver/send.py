"""`uv run receiver-send`: send one signed delivery, as Hookline's worker will (stdlib only)."""

import argparse
import time
import urllib.error
import urllib.request
import uuid
from collections.abc import Sequence

from receiver.signature import sign


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(prog="receiver-send")
    parser.add_argument("url", help="for example http://127.0.0.1:8081/hooks/orders")
    parser.add_argument("--secret", required=True)
    parser.add_argument("--body", default='{"type":"invoice.paid","data":{"amount_cents":1999}}')
    parser.add_argument("--event-id", default=None, help="default: a new random id")
    parser.add_argument("--skew", type=int, default=0, help="shift the timestamp by N seconds")
    args = parser.parse_args(argv)

    body = args.body.encode()
    timestamp = int(time.time()) + args.skew
    headers = {
        "Content-Type": "application/json",
        "Hookline-Event-Id": args.event_id or f"evt_{uuid.uuid4().hex[:12]}",
        "Hookline-Delivery-Id": f"dlv_{uuid.uuid4().hex[:12]}",
        "Hookline-Attempt": "1",
        "Hookline-Signature": sign(args.secret.encode(), timestamp, body),
    }
    request = urllib.request.Request(args.url, data=body, headers=headers, method="POST")
    try:
        with urllib.request.urlopen(request, timeout=10) as response:
            print(response.status, response.read().decode())
            return 0
    except urllib.error.HTTPError as error:
        print(error.code, error.read().decode())
        return 1
