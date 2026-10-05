# The receiver simulator (chapter 8) as a container, for the Hookline compose stack. Milestone M7
# hardens it (non-root user, smaller image, SBOM); this version only has to run.
FROM python:3.14.8-slim
COPY --from=ghcr.io/astral-sh/uv:0.12.23 /uv /usr/local/bin/uv
WORKDIR /app
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy UV_PROJECT_ENVIRONMENT=/app/.venv
# Dependencies first, so editing the code does not reinstall them.
COPY pyproject.toml uv.lock .python-version ./
RUN uv sync --frozen --no-dev --no-install-project
COPY src ./src
RUN uv sync --frozen --no-dev
ENV PATH=/app/.venv/bin:$PATH RECEIVER_HOST=0.0.0.0 RECEIVER_PORT=8081
EXPOSE 8081
# One worker: the receipts and counters live in this process's memory (chapter 8, section 8.7).
CMD ["sh", "-c", "exec uvicorn --factory receiver.app:create_app --host \"$RECEIVER_HOST\" --port \"$RECEIVER_PORT\" --workers 1"]
