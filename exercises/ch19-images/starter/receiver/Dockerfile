# syntax=docker/dockerfile:1
# The chapter 8 receiver on python slim, installed with uv. Build from this
# directory: docker build -t receiver:dev .

# ---- build stage ----
FROM python:3.14.8-slim AS build
# TODO(day 58): copy the uv binary from ghcr.io/astral-sh/uv:0.12.23 to /bin/uv.
# TODO(day 58): set UV_COMPILE_BYTECODE=1, UV_LINK_MODE=copy, UV_PYTHON_DOWNLOADS=0.
WORKDIR /app

# TODO(day 58): install the dependencies only (uv sync --locked --no-dev
# --no-install-project), with a cache mount on /root/.cache/uv and bind mounts
# for uv.lock and pyproject.toml, so the layer depends on the lock file alone.

# TODO(day 58): copy the project and install it, non-editable, into the same
# environment (uv sync --locked --no-dev --no-editable).

# ---- runtime stage ----
FROM python:3.14.8-slim
# TODO(day 58): a GIT_SHA build argument and the revision label, as in shorty.
# TODO(day 59, after the scan): remove the base image's pip.
# TODO(day 58): create the group and user app with id 10001 (no home, no shell).
# TODO(day 58): copy /app/.venv from the build stage; put /app/.venv/bin first
# on PATH; set PYTHONDONTWRITEBYTECODE=1, PYTHONUNBUFFERED=1,
# RECEIVER_HOST=0.0.0.0 and RECEIVER_PORT=8081.
# TODO(day 58): run as 10001:10001, expose 8081, entrypoint "receiver".
