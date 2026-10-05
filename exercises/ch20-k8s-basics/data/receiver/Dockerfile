# syntax=docker/dockerfile:1
# The chapter 8 receiver on python slim, installed with uv. Build from this
# directory: docker build -t receiver:dev .

# ---- build stage: uv resolves and installs into /app/.venv ----
FROM python:3.14.8-slim AS build
COPY --from=ghcr.io/astral-sh/uv:0.12.23 /uv /bin/uv
ENV UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy \
    UV_PYTHON_DOWNLOADS=0
WORKDIR /app

# Dependencies first, from the lock file only: cached until uv.lock changes.
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv sync --locked --no-dev --no-install-project

# Then the project itself, installed as a regular (non-editable) package.
COPY . .
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev --no-editable

# ---- runtime stage: the same python, the virtual environment, no uv ----
FROM python:3.14.8-slim
ARG GIT_SHA=unknown
LABEL org.opencontainers.image.revision="${GIT_SHA}"
# The app never installs packages at runtime, so the base image's pip goes: its own
# package metadata is what the scanners reported. See section 19.6.
RUN rm -rf /usr/local/lib/python3.14/site-packages/pip* /usr/local/bin/pip* \
 && groupadd --system --gid 10001 app \
 && useradd --system --uid 10001 --gid app --no-create-home --shell /usr/sbin/nologin app
# Owned by root and not writable by the app user: the code cannot be changed at runtime.
COPY --from=build /app/.venv /app/.venv
ENV PATH="/app/.venv/bin:${PATH}" \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    RECEIVER_HOST=0.0.0.0 \
    RECEIVER_PORT=8081
USER 10001:10001
EXPOSE 8081
ENTRYPOINT ["receiver"]
