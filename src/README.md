README.md patch content

Add a Python install section like this:

## Python install

This repository is installable as a Python package from Git:

```bash
python -m pip install "vnodex @ git+https://github.com/YOUR_ORG/v-nodex.git@main"

Example usage:

from vnodex import VNodeXClient

client = VNodeXClient("http://127.0.0.1:8080")
print(client.health().status)