import socket
import time
from datetime import datetime
import logging

logging.basicConfig(level=logging.INFO)


def new_socket(port: int) -> socket.socket:
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEPORT, 1)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_BROADCAST, 1)
    s.settimeout(1)
    s.bind(("", port))
    return s


port = 8080
timeFmt = "%H:%M:%S"
s = new_socket(port)

logging.info("new server created at " + datetime.now().strftime(timeFmt))

while True:
    fmt = datetime.now().strftime(timeFmt)
    s.sendto(fmt.encode("utf-8"), ("<broadcast>", port))
    logging.info("broadcasting " + fmt)
    time.sleep(1)
