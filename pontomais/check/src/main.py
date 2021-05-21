#!/usr/bin/python3
import os
from formula import formula

email = os.environ.get("RIT_INPUT_EMAIL")
password = os.environ.get("RIT_INPUT_PASSWORD")
formula.Run(email, password)