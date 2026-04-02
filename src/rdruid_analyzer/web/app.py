from pathlib import Path

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse

from rdruid_analyzer.web.routes import router


def create_app() -> FastAPI:
    app = FastAPI(title="Resto Druid Talent Analyzer")
    app.include_router(router)

    static_dir = Path(__file__).parent.parent.parent.parent / "frontend" / "dist"
    if static_dir.exists():
        app.mount("/assets", StaticFiles(directory=static_dir / "assets"), name="assets")

        @app.get("/{path:path}")
        async def spa_fallback(path: str):
            return FileResponse(static_dir / "index.html")

    return app


app = create_app()
