"""count.py - read the answer to the gate's count query, or say what is wrong with it.

Usage: python3 -I count.py <file holding the response body> <the column name the query asked for>

Prints the count and exits 0 only when the answer is exactly:
  a JSON object, one document and nothing after it,
  with "headers" a list of one name equal to the column asked for (case aside),
  with "data" a list of exactly one row of exactly one cell,
  the cell a non-negative whole number (a JSON integer, or a string of digits),
  and no field named error, errors, exception, message or msg.
Anything else exits 1 and prints, on stderr, the one thing that was wrong. No answer is never a pass.
"""
import json
import sys

REFUSING_FIELDS = ("error", "errors", "exception", "message", "msg")


def fail(why):
    print(why, file=sys.stderr)
    sys.exit(1)


def main():
    if len(sys.argv) != 3:
        fail("usage: count.py <body file> <column>")
    path, column = sys.argv[1], sys.argv[2]
    try:
        with open(path, encoding="utf-8") as f:
            doc = json.load(f)
    except (OSError, ValueError) as e:
        fail("the body is not one JSON document (%s)" % type(e).__name__)
    if not isinstance(doc, dict):
        fail("the body is not a JSON object")
    for field in doc:
        if field.lower() in REFUSING_FIELDS:
            fail("the body carries a '%s' field" % field.lower())
    headers = doc.get("headers")
    if not isinstance(headers, list) or len(headers) != 1 or not isinstance(headers[0], str) or headers[0].lower() != column.lower():
        fail("'headers' is not exactly [\"%s\"]" % column)
    data = doc.get("data")
    if not isinstance(data, list) or len(data) != 1:
        fail("'data' is not exactly one row")
    row = data[0]
    if not isinstance(row, list) or len(row) != 1:
        fail("the row is not exactly one cell")
    cell = row[0]
    if isinstance(cell, bool):
        fail("the cell is not a number")
    if isinstance(cell, str) and cell.isascii() and cell.isdigit():
        cell = int(cell)
    if not isinstance(cell, int) or cell < 0:
        fail("the cell is not a non-negative whole number")
    print(cell)


main()
