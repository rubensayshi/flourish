# Web App Implementation Plan

**Goal:** Add a web frontend (FastAPI + Vue) so users can analyze WCL reports in the browser without installing anything.

**Architecture:** FastAPI backend serves a Vue 3 SPA and two API endpoints. Single Fly.io service with disk-based caching. Existing analysis engine is reused as-is.

**Tech Stack:** FastAPI, uvicorn, Vue 3, Vite, Tailwind CSS, vue-router

---

## File Structure

```
src/rdruid_analyzer/
  web/
    __init__.py
    app.py           # FastAPI app, static mount, catch-all
    routes.py        # API endpoints
    cache.py         # Disk-based analysis result cache
    dependencies.py  # Shared FastAPI deps (WCL client singleton)

frontend/
  package.json
  vite.config.js
  tailwind.config.js
  postcss.config.js
  index.html
  src/
    main.js
    App.vue
    router.js
    api.js           # API client helper
    views/
      Home.vue
      Analyze.vue
    components/
      ReportInput.vue
      FightSelector.vue
      PlayerSelector.vue
      ResultsTable.vue

tests/
  web/
    __init__.py
    test_routes.py
    test_cache.py

Dockerfile
fly.toml
```

---

## Task 1: Backend — Result Cache

**Files:**
- Create: `src/rdruid_analyzer/web/__init__.py`
- Create: `src/rdruid_analyzer/web/cache.py`
- Create: `tests/web/__init__.py`
- Create: `tests/web/test_cache.py`

- [ ] **Step 1: Write failing test**

```python
# tests/web/test_cache.py
import json
from rdruid_analyzer.web.cache import ResultCache


def test_get_returns_none_when_missing(tmp_path):
    cache = ResultCache(cache_dir=tmp_path)
    assert cache.get("ABC123", 1, "Player") is None


def test_set_then_get_roundtrips(tmp_path):
    cache = ResultCache(cache_dir=tmp_path)
    data = {"total_healing": 1000, "talents": [{"name": "SotF", "attributed": 500}]}
    cache.set("ABC123", 1, "Player", data)
    assert cache.get("ABC123", 1, "Player") == data


def test_cache_key_is_case_insensitive_for_player(tmp_path):
    cache = ResultCache(cache_dir=tmp_path)
    data = {"total_healing": 1000}
    cache.set("ABC123", 1, "Saikó", data)
    assert cache.get("ABC123", 1, "saikó") == data
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `uv run pytest tests/web/test_cache.py -v`
Expected: ImportError — module doesn't exist yet

- [ ] **Step 3: Implement**

```python
# src/rdruid_analyzer/web/__init__.py
# (empty)
```

```python
# tests/web/__init__.py
# (empty)
```

```python
# src/rdruid_analyzer/web/cache.py
import json
from pathlib import Path


class ResultCache:
    def __init__(self, cache_dir: Path = Path("data/results_cache")):
        self._dir = cache_dir
        self._dir.mkdir(parents=True, exist_ok=True)

    def _key_path(self, code: str, fight_id: int, player: str) -> Path:
        safe_player = player.lower().replace(" ", "_")
        return self._dir / f"{code}_{fight_id}_{safe_player}.json"

    def get(self, code: str, fight_id: int, player: str) -> dict | None:
        path = self._key_path(code, fight_id, player)
        if not path.exists():
            return None
        return json.loads(path.read_text())

    def set(self, code: str, fight_id: int, player: str, data: dict) -> None:
        path = self._key_path(code, fight_id, player)
        path.write_text(json.dumps(data))
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `uv run pytest tests/web/test_cache.py -v`
Expected: 3 passed

- [ ] **Step 5: Commit**

```bash
git add src/rdruid_analyzer/web/ tests/web/
git commit -m "feat(web): add disk-based result cache"
```

---

## Task 2: Backend — API Routes

**Files:**
- Create: `src/rdruid_analyzer/web/dependencies.py`
- Create: `src/rdruid_analyzer/web/routes.py`
- Create: `tests/web/test_routes.py`

- [ ] **Step 1: Write failing tests**

```python
# tests/web/test_routes.py
from unittest.mock import patch, MagicMock
from fastapi.testclient import TestClient

from rdruid_analyzer.web.app import create_app


def _mock_wcl_client():
    client = MagicMock()
    client.get_report.return_value = {
        "title": "Test Raid",
        "fights": [
            {"id": 1, "name": "Boss", "kill": True, "startTime": 0,
             "endTime": 60000, "encounterID": 123, "difficulty": 4},
            {"id": 2, "name": "Trash", "kill": True, "startTime": 0,
             "endTime": 30000, "encounterID": 0, "difficulty": 0},
        ],
        "masterData": {
            "actors": [
                {"id": 10, "name": "Saikó", "type": "Player",
                 "subType": "Druid", "server": "Draenor", "petOwner": None},
                {"id": 11, "name": "Warrior", "type": "Player",
                 "subType": "Warrior", "server": "Draenor", "petOwner": None},
            ]
        },
    }
    return client


@patch("rdruid_analyzer.web.dependencies.get_wcl_client")
def test_report_endpoint_returns_fights_and_druids(mock_get_client):
    mock_get_client.return_value = _mock_wcl_client()
    app = create_app()
    client = TestClient(app)

    resp = client.get("/api/report/ABC123")
    assert resp.status_code == 200
    data = resp.json()
    assert data["title"] == "Test Raid"
    # Only boss fights (encounterID > 0)
    assert len(data["fights"]) == 1
    assert data["fights"][0]["name"] == "Boss"
    # Only druids
    assert len(data["druids"]) == 1
    assert data["druids"][0]["name"] == "Saikó"


@patch("rdruid_analyzer.web.dependencies.get_wcl_client")
def test_report_endpoint_404_on_invalid_code(mock_get_client):
    mock_get_client.return_value = MagicMock(
        get_report=MagicMock(side_effect=Exception("not found"))
    )
    app = create_app()
    client = TestClient(app)

    resp = client.get("/api/report/INVALID")
    assert resp.status_code == 404


def test_health_endpoint():
    app = create_app()
    client = TestClient(app)
    resp = client.get("/api/health")
    assert resp.status_code == 200
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `uv run pytest tests/web/test_routes.py -v`
Expected: ImportError — modules don't exist

- [ ] **Step 3: Add fastapi and uvicorn to dependencies**

Add to `pyproject.toml` dependencies:

```toml
dependencies = [
    "httpx>=0.27",
    "typer>=0.12",
    "rich>=13",
    "pyyaml>=6",
    "python-dotenv>=1",
    "fastapi>=0.115",
    "uvicorn>=0.34",
]
```

Run: `uv sync --all-extras`

- [ ] **Step 4: Implement dependencies.py**

```python
# src/rdruid_analyzer/web/dependencies.py
import os
from functools import lru_cache

from dotenv import load_dotenv

from rdruid_analyzer.wcl.client import WCLClient
from rdruid_analyzer.wcl.cache import CachedWCLClient


@lru_cache
def get_wcl_client() -> CachedWCLClient:
    load_dotenv()
    client_id = os.environ.get("WCL_CLIENT_ID", "")
    client_secret = os.environ.get("WCL_CLIENT_SECRET", "")
    if not client_id or not client_secret:
        raise RuntimeError("WCL_CLIENT_ID and WCL_CLIENT_SECRET must be set")
    return CachedWCLClient(WCLClient(client_id, client_secret))
```

- [ ] **Step 5: Implement routes.py**

```python
# src/rdruid_analyzer/web/routes.py
from fastapi import APIRouter, HTTPException

from rdruid_analyzer.web.dependencies import get_wcl_client
from rdruid_analyzer.web.cache import ResultCache
from rdruid_analyzer.models.config import load_config, Config, MasteryConfig
from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.cli import build_attributors

import os

router = APIRouter(prefix="/api")
result_cache = ResultCache()

DRUID_CLASS = "Druid"


@router.get("/health")
def health():
    return {"status": "ok"}


@router.get("/report/{code}")
def get_report(code: str):
    client = get_wcl_client()
    try:
        report = client.get_report(code)
    except Exception:
        raise HTTPException(status_code=404, detail="Report not found")

    fights = [
        {
            "id": f["id"],
            "name": f["name"],
            "kill": f["kill"],
            "duration": round((f["endTime"] - f["startTime"]) / 1000),
        }
        for f in report["fights"]
        if f.get("encounterID", 0) > 0
    ]

    druids = [
        {"id": a["id"], "name": a["name"], "server": a.get("server", "")}
        for a in report["masterData"]["actors"]
        if a.get("subType") == DRUID_CLASS
    ]

    return {"title": report["title"], "fights": fights, "druids": druids}


@router.get("/analyze/{code}/{fight_id}/{player_name}")
def analyze(code: str, fight_id: int, player_name: str):
    # Check result cache first
    cached = result_cache.get(code, fight_id, player_name)
    if cached:
        return cached

    client = get_wcl_client()
    try:
        report = client.get_report(code)
    except Exception:
        raise HTTPException(status_code=404, detail="Report not found")

    selected_fight = next(
        (f for f in report["fights"] if f["id"] == fight_id), None
    )
    if not selected_fight:
        raise HTTPException(status_code=404, detail="Fight not found")

    all_actors = report["masterData"]["actors"]
    selected_player = next(
        (a for a in all_actors
         if a.get("subType") == DRUID_CLASS
         and a["name"].lower() == player_name.lower()),
        None,
    )
    if not selected_player:
        raise HTTPException(status_code=404, detail="Player not found")

    raw_events = client.get_events(
        code, fight_id, selected_player["id"],
        selected_fight["startTime"], selected_fight["endTime"],
    )

    REGROWTH_BUFF_ID = 8936
    regrowth_filter = (
        f'IN RANGE FROM (type = "applybuff" OR type = "refreshbuff") '
        f"AND ability.id = {REGROWTH_BUFF_ID} "
        f'TO type = "removebuff" AND ability.id = {REGROWTH_BUFF_ID} '
        f"GROUP BY target ON target END"
    )
    damage_taken_with_regrowth = client.get_damage_taken(
        code, fight_id, selected_player["id"],
        selected_fight["startTime"], selected_fight["endTime"],
        filter_expression=regrowth_filter,
    )

    config_path = "config/talents.yaml"
    config = (
        load_config(config_path)
        if os.path.exists(config_path)
        else Config(mastery=MasteryConfig(), talents={})
    )

    pet_ids = {a["id"] for a in all_actors if a.get("petOwner")}
    attributors = build_attributors(config, damage_taken_with_regrowth=damage_taken_with_regrowth)
    pipeline = Pipeline(attributors=attributors, pet_ids=pet_ids)
    results = pipeline.run(raw_events)

    duration_sec = max(results.fight_duration_ms / 1000, 1)
    talents = [
        {
            "name": name,
            "attributed": round(amount),
            "pct": round(amount / results.total_healing * 100, 1) if results.total_healing > 0 else 0,
            "hps": round(amount / duration_sec),
        }
        for name, amount in sorted(
            results.talent_healing.items(), key=lambda x: x[1], reverse=True
        )
        if amount > 0
    ]

    total_attributed = sum(t["attributed"] for t in talents)
    unattributed = max(0, results.total_healing - total_attributed - results.wasted)

    response = {
        "fight_name": selected_fight["name"],
        "player_name": selected_player["name"],
        "total_healing": results.total_healing,
        "duration_sec": round(duration_sec),
        "talents": talents,
        "wasted": results.wasted,
        "unattributed": unattributed,
    }

    result_cache.set(code, fight_id, player_name, response)
    return response
```

- [ ] **Step 6: Implement app.py**

```python
# src/rdruid_analyzer/web/app.py
from pathlib import Path

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse

from rdruid_analyzer.web.routes import router


def create_app() -> FastAPI:
    app = FastAPI(title="Resto Druid Talent Analyzer")
    app.include_router(router)

    # Serve Vue SPA static files if built
    static_dir = Path(__file__).parent.parent.parent.parent / "frontend" / "dist"
    if static_dir.exists():
        app.mount("/assets", StaticFiles(directory=static_dir / "assets"), name="assets")

        @app.get("/{path:path}")
        async def spa_fallback(path: str):
            return FileResponse(static_dir / "index.html")

    return app


app = create_app()
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `uv run pytest tests/web/ -v`
Expected: All pass

- [ ] **Step 8: Commit**

```bash
git add src/rdruid_analyzer/web/ tests/web/ pyproject.toml
git commit -m "feat(web): add FastAPI backend with report and analyze endpoints"
```

---

## Task 3: Backend — Rate Limiting

**Files:**
- Modify: `src/rdruid_analyzer/web/app.py`

- [ ] **Step 1: Add slowapi to dependencies**

Add to `pyproject.toml`:
```
"slowapi>=0.1.9",
```

Run: `uv sync --all-extras`

- [ ] **Step 2: Add rate limiting to app.py**

Update `create_app()` in `app.py`:

```python
from slowapi import Limiter
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded
from slowapi.middleware import SlowAPIMiddleware
from fastapi.responses import JSONResponse

limiter = Limiter(key_func=get_remote_address)


def create_app() -> FastAPI:
    app = FastAPI(title="Resto Druid Talent Analyzer")
    app.state.limiter = limiter
    app.add_middleware(SlowAPIMiddleware)

    @app.exception_handler(RateLimitExceeded)
    async def rate_limit_handler(request, exc):
        return JSONResponse(
            status_code=429,
            content={"detail": "Rate limit exceeded. Try again in a minute."},
        )

    app.include_router(router)

    # ... static files mount unchanged
```

Then in `routes.py`, decorate the analyze endpoint:

```python
from fastapi import Request
from slowapi import Limiter
from slowapi.util import get_remote_address

limiter = Limiter(key_func=get_remote_address)

@router.get("/analyze/{code}/{fight_id}/{player_name}")
@limiter.limit("10/minute")
def analyze(request: Request, code: str, fight_id: int, player_name: str):
    # ... unchanged
```

And in `app.py`, import the limiter from routes instead of defining it locally:

```python
from rdruid_analyzer.web.routes import router, limiter
```

- [ ] **Step 3: Manual test**

Run: `uv run uvicorn rdruid_analyzer.web.app:app --port 8080`
Hit `http://localhost:8080/api/health` — should return `{"status": "ok"}`

- [ ] **Step 4: Commit**

```bash
git add pyproject.toml src/rdruid_analyzer/web/
git commit -m "feat(web): add IP-based rate limiting on analyze endpoint"
```

---

## Task 4: Frontend — Project Scaffold

**Files:**
- Create: `frontend/` (Vue + Vite + Tailwind scaffold)

- [ ] **Step 1: Scaffold Vue project**

```bash
cd frontend
npm create vite@latest . -- --template vue
npm install
npm install -D tailwindcss @tailwindcss/vite
npm install vue-router
```

- [ ] **Step 2: Configure Tailwind**

```js
// frontend/vite.config.js
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
```

```css
/* frontend/src/style.css */
@import "tailwindcss";
```

- [ ] **Step 3: Verify it builds**

```bash
cd frontend && npm run build
```

Expected: `frontend/dist/` created with `index.html` and `assets/`

- [ ] **Step 4: Commit**

```bash
git add frontend/
git commit -m "feat(web): scaffold Vue + Vite + Tailwind frontend"
```

---

## Task 5: Frontend — Router + API Client

**Files:**
- Create: `frontend/src/router.js`
- Create: `frontend/src/api.js`
- Modify: `frontend/src/main.js`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Create router**

```js
// frontend/src/router.js
import { createRouter, createWebHistory } from 'vue-router'
import Home from './views/Home.vue'
import Analyze from './views/Analyze.vue'

const routes = [
  { path: '/', component: Home },
  { path: '/analyze/:code', component: Analyze },
  { path: '/results/:code/:fightId/:player', component: Analyze },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
```

- [ ] **Step 2: Create API client**

```js
// frontend/src/api.js
const API_BASE = '/api'

export async function fetchReport(code) {
  const res = await fetch(`${API_BASE}/report/${code}`)
  if (!res.ok) throw new Error(`Report not found (${res.status})`)
  return res.json()
}

export async function fetchAnalysis(code, fightId, player) {
  const res = await fetch(`${API_BASE}/analyze/${code}/${fightId}/${encodeURIComponent(player)}`)
  if (!res.ok) {
    if (res.status === 429) throw new Error('Rate limit exceeded. Try again in a minute.')
    throw new Error(`Analysis failed (${res.status})`)
  }
  return res.json()
}
```

- [ ] **Step 3: Wire up main.js and App.vue**

```js
// frontend/src/main.js
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './style.css'

createApp(App).use(router).mount('#app')
```

```vue
<!-- frontend/src/App.vue -->
<template>
  <div class="min-h-screen bg-slate-900 text-slate-100">
    <header class="border-b border-slate-700 px-6 py-4">
      <router-link to="/" class="text-xl font-bold text-emerald-400 hover:text-emerald-300">
        Resto Druid Talent Analyzer
      </router-link>
    </header>
    <main class="mx-auto max-w-3xl px-4 py-8">
      <router-view />
    </main>
  </div>
</template>
```

- [ ] **Step 4: Create placeholder views**

```vue
<!-- frontend/src/views/Home.vue -->
<template>
  <div>Home (placeholder)</div>
</template>
```

```vue
<!-- frontend/src/views/Analyze.vue -->
<template>
  <div>Analyze (placeholder)</div>
</template>
```

- [ ] **Step 5: Verify dev server works**

Run backend: `uv run uvicorn rdruid_analyzer.web.app:app --port 8080`
Run frontend: `cd frontend && npm run dev`

Navigate to `http://localhost:5173/` — should show header + "Home (placeholder)"

- [ ] **Step 6: Commit**

```bash
git add frontend/
git commit -m "feat(web): add vue-router and API client"
```

---

## Task 6: Frontend — Home View (Report Input)

**Files:**
- Create: `frontend/src/components/ReportInput.vue`
- Modify: `frontend/src/views/Home.vue`

- [ ] **Step 1: Implement ReportInput.vue**

```vue
<!-- frontend/src/components/ReportInput.vue -->
<template>
  <form @submit.prevent="submit" class="flex gap-3">
    <input
      v-model="input"
      type="text"
      placeholder="Paste WarcraftLogs URL or report code..."
      class="flex-1 rounded-lg bg-slate-800 border border-slate-600 px-4 py-3
             text-slate-100 placeholder-slate-500
             focus:outline-none focus:border-emerald-500"
    />
    <button
      type="submit"
      :disabled="!code"
      class="rounded-lg bg-emerald-600 px-6 py-3 font-semibold text-white
             hover:bg-emerald-500 disabled:opacity-40 disabled:cursor-not-allowed"
    >
      Analyze
    </button>
  </form>
  <p v-if="input && !code" class="mt-2 text-sm text-red-400">
    Could not extract a report code from this input.
  </p>
</template>

<script setup>
import { ref, computed } from 'vue'

const emit = defineEmits(['submit'])
const input = ref('')

const code = computed(() => {
  const text = input.value.trim()
  // Match full URL: warcraftlogs.com/reports/CODE or just the code
  const match = text.match(/warcraftlogs\.com\/reports\/([A-Za-z0-9]+)/)
  if (match) return match[1]
  // Bare code (alphanumeric, 16 chars)
  if (/^[A-Za-z0-9]{10,20}$/.test(text)) return text
  return null
})

function submit() {
  if (code.value) emit('submit', code.value)
}
</script>
```

- [ ] **Step 2: Implement Home.vue**

```vue
<!-- frontend/src/views/Home.vue -->
<template>
  <div class="mt-16 text-center">
    <h1 class="text-3xl font-bold text-emerald-400 mb-2">
      Resto Druid Talent Analyzer
    </h1>
    <p class="text-slate-400 mb-8">
      See how much healing each talent contributes in your log.
    </p>
    <div class="max-w-xl mx-auto">
      <ReportInput @submit="goToReport" />
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import ReportInput from '../components/ReportInput.vue'

const router = useRouter()

function goToReport(code) {
  router.push(`/analyze/${code}`)
}
</script>
```

- [ ] **Step 3: Verify in browser**

Paste `https://www.warcraftlogs.com/reports/ZMq6Axh2z9N37RWC` — should extract code and navigate to `/analyze/ZMq6Axh2z9N37RWC`.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/
git commit -m "feat(web): add landing page with report URL input"
```

---

## Task 7: Frontend — Analyze View (Selectors)

**Files:**
- Create: `frontend/src/components/FightSelector.vue`
- Create: `frontend/src/components/PlayerSelector.vue`
- Create: `frontend/src/components/LoadingSpinner.vue`
- Modify: `frontend/src/views/Analyze.vue`

- [ ] **Step 1: Implement LoadingSpinner.vue**

```vue
<!-- frontend/src/components/LoadingSpinner.vue -->
<template>
  <div class="flex items-center gap-3 text-slate-400">
    <svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
      <path class="opacity-75" fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
    <span><slot>Loading...</slot></span>
  </div>
</template>
```

- [ ] **Step 2: Implement FightSelector.vue**

```vue
<!-- frontend/src/components/FightSelector.vue -->
<template>
  <div>
    <label class="block text-sm text-slate-400 mb-1">Fight</label>
    <select
      :value="modelValue"
      @change="$emit('update:modelValue', Number($event.target.value))"
      class="w-full rounded-lg bg-slate-800 border border-slate-600 px-4 py-2.5
             text-slate-100 focus:outline-none focus:border-emerald-500"
    >
      <option :value="0" disabled>Select a fight...</option>
      <option v-for="f in fights" :key="f.id" :value="f.id">
        {{ f.name }} — {{ f.kill ? 'Kill' : 'Wipe' }} ({{ f.duration }}s)
      </option>
    </select>
  </div>
</template>

<script setup>
defineProps({
  fights: Array,
  modelValue: Number,
})
defineEmits(['update:modelValue'])
</script>
```

- [ ] **Step 3: Implement PlayerSelector.vue**

```vue
<!-- frontend/src/components/PlayerSelector.vue -->
<template>
  <div>
    <label class="block text-sm text-slate-400 mb-1">Player</label>
    <select
      :value="modelValue"
      @change="$emit('update:modelValue', $event.target.value)"
      class="w-full rounded-lg bg-slate-800 border border-slate-600 px-4 py-2.5
             text-slate-100 focus:outline-none focus:border-emerald-500"
    >
      <option value="" disabled>Select a druid...</option>
      <option v-for="d in druids" :key="d.id" :value="d.name">
        {{ d.name }} ({{ d.server }})
      </option>
    </select>
  </div>
</template>

<script setup>
defineProps({
  druids: Array,
  modelValue: String,
})
defineEmits(['update:modelValue'])
</script>
```

- [ ] **Step 4: Implement Analyze.vue**

```vue
<!-- frontend/src/views/Analyze.vue -->
<template>
  <div>
    <!-- Loading report -->
    <LoadingSpinner v-if="loading">Loading report...</LoadingSpinner>

    <!-- Error -->
    <div v-else-if="error" class="text-red-400">{{ error }}</div>

    <!-- Report loaded -->
    <template v-else-if="report">
      <h2 class="text-xl font-bold mb-4">{{ report.title }}</h2>

      <div class="grid grid-cols-2 gap-4 mb-6">
        <FightSelector v-model="selectedFight" :fights="report.fights" />
        <PlayerSelector v-model="selectedPlayer" :druids="report.druids" />
      </div>

      <button
        @click="runAnalysis"
        :disabled="!selectedFight || !selectedPlayer || analyzing"
        class="rounded-lg bg-emerald-600 px-6 py-2.5 font-semibold text-white
               hover:bg-emerald-500 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        {{ analyzing ? 'Analyzing...' : 'Run Analysis' }}
      </button>

      <LoadingSpinner v-if="analyzing" class="mt-4">
        Analyzing (this may take a few seconds)...
      </LoadingSpinner>

      <ResultsTable v-if="results" :data="results" class="mt-6" />
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchReport, fetchAnalysis } from '../api'
import FightSelector from '../components/FightSelector.vue'
import PlayerSelector from '../components/PlayerSelector.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import ResultsTable from '../components/ResultsTable.vue'

const route = useRoute()
const router = useRouter()

const report = ref(null)
const loading = ref(true)
const error = ref(null)
const selectedFight = ref(0)
const selectedPlayer = ref('')
const analyzing = ref(false)
const results = ref(null)

onMounted(async () => {
  try {
    report.value = await fetchReport(route.params.code)

    // If permalink route, auto-select and run
    if (route.params.fightId && route.params.player) {
      selectedFight.value = Number(route.params.fightId)
      selectedPlayer.value = route.params.player
      await runAnalysis()
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

async function runAnalysis() {
  analyzing.value = true
  error.value = null
  try {
    results.value = await fetchAnalysis(
      route.params.code, selectedFight.value, selectedPlayer.value
    )
    router.replace(`/results/${route.params.code}/${selectedFight.value}/${selectedPlayer.value}`)
  } catch (e) {
    error.value = e.message
  } finally {
    analyzing.value = false
  }
}
</script>
```

- [ ] **Step 5: Verify in browser**

Navigate to `/analyze/ZMq6Axh2z9N37RWC` (with backend running). Should show report title, fight dropdown, player dropdown.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/
git commit -m "feat(web): add fight/player selection on analyze view"
```

---

## Task 8: Frontend — Results Table

**Files:**
- Create: `frontend/src/components/ResultsTable.vue`

- [ ] **Step 1: Implement ResultsTable.vue**

```vue
<!-- frontend/src/components/ResultsTable.vue -->
<template>
  <div>
    <div class="mb-4 text-slate-400">
      <span class="font-bold text-slate-100">{{ data.fight_name }}</span>
      &mdash;
      <span class="font-bold text-slate-100">{{ data.player_name }}</span>
      &mdash;
      Total healing: <span class="text-emerald-400 font-bold">{{ fmt(data.total_healing) }}</span>
      ({{ data.duration_sec }}s)
    </div>

    <table class="w-full text-sm">
      <thead>
        <tr class="text-left text-slate-400 border-b border-slate-700">
          <th class="py-2 pr-4">Talent</th>
          <th class="py-2 pr-4 text-right">Attributed</th>
          <th class="py-2 pr-4 text-right">% Total</th>
          <th class="py-2 text-right">HPS</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(t, i) in data.talents"
          :key="t.name"
          :class="i % 2 === 0 ? 'bg-slate-800/50' : ''"
          class="border-b border-slate-800"
        >
          <td class="py-1.5 pr-4">{{ t.name }}</td>
          <td class="py-1.5 pr-4 text-right font-mono">{{ fmt(t.attributed) }}</td>
          <td class="py-1.5 pr-4 text-right font-mono" :class="pctColor(t.pct)">
            {{ t.pct.toFixed(1) }}%
          </td>
          <td class="py-1.5 text-right font-mono">{{ fmt(t.hps) }}</td>
        </tr>
      </tbody>
      <tfoot class="text-slate-500">
        <tr class="border-t border-slate-700">
          <td class="py-1.5 pr-4">Wasted (&gt;50% OH)</td>
          <td class="py-1.5 pr-4 text-right font-mono">{{ fmt(data.wasted) }}</td>
          <td class="py-1.5 pr-4 text-right">&mdash;</td>
          <td class="py-1.5 text-right">&mdash;</td>
        </tr>
        <tr>
          <td class="py-1.5 pr-4">Unattributed</td>
          <td class="py-1.5 pr-4 text-right font-mono">{{ fmt(data.unattributed) }}</td>
          <td class="py-1.5 pr-4 text-right">&mdash;</td>
          <td class="py-1.5 text-right">&mdash;</td>
        </tr>
      </tfoot>
    </table>

    <p v-if="totalAttributed > data.total_healing" class="mt-4 text-xs text-slate-500">
      Talents can overlap (multiple talents buff the same heal).
      Total attributed may exceed total healing.
    </p>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({ data: Object })

const totalAttributed = computed(() =>
  props.data.talents.reduce((sum, t) => sum + t.attributed, 0)
)

function fmt(n) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'k'
  return String(Math.round(n))
}

function pctColor(pct) {
  if (pct >= 5) return 'text-emerald-400'
  if (pct >= 2) return 'text-emerald-600'
  return 'text-slate-400'
}
</script>
```

- [ ] **Step 2: Verify end-to-end**

With backend + frontend dev servers running:
1. Go to `/`
2. Paste a WCL URL
3. Select fight + player
4. Click "Run Analysis"
5. Verify table renders with correct data
6. Verify URL updates to `/results/...` and is shareable

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ResultsTable.vue
git commit -m "feat(web): add results table component"
```

---

## Task 9: Deployment — Dockerfile + fly.toml

**Files:**
- Create: `Dockerfile`
- Create: `fly.toml`
- Modify: `.gitignore`

- [ ] **Step 1: Create Dockerfile**

```dockerfile
# Dockerfile

# Stage 1: Build Vue frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Python app
FROM python:3.12-slim
WORKDIR /app

# Install uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

# Install Python deps
COPY pyproject.toml uv.lock ./
RUN uv sync --no-dev --frozen

# Copy app code
COPY src/ src/
COPY config/ config/

# Copy built frontend
COPY --from=frontend /app/frontend/dist frontend/dist

EXPOSE 8080
CMD ["uv", "run", "uvicorn", "rdruid_analyzer.web.app:app", "--host", "0.0.0.0", "--port", "8080"]
```

- [ ] **Step 2: Create fly.toml**

```toml
# fly.toml
app = "rdruid-talent-analyzer"
primary_region = "ams"

[build]

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = "stop"
  auto_start_machines = true
  min_machines_running = 0

[mounts]
  source = "data"
  destination = "/data"

[[vm]]
  size = "shared-cpu-1x"
  memory = "256mb"

[checks]
  [checks.health]
    type = "http"
    port = 8080
    path = "/api/health"
    interval = "30s"
    timeout = "5s"
```

- [ ] **Step 3: Update .gitignore**

Add:
```
frontend/node_modules/
frontend/dist/
```

- [ ] **Step 4: Update cache paths for Fly volume**

In `src/rdruid_analyzer/web/cache.py`, update default:
```python
class ResultCache:
    def __init__(self, cache_dir: Path = Path("/data/results_cache")):
```

In `src/rdruid_analyzer/wcl/cache.py`, check if `/data` exists for Fly:
```python
DEFAULT_CACHE_DIR = Path("/data/wcl_cache") if Path("/data").exists() else Path("data/cache")
```

- [ ] **Step 5: Verify Docker build**

```bash
docker build -t rdruid-web .
docker run -p 8080:8080 --env-file .env rdruid-web
```

Hit `http://localhost:8080/api/health` — should return 200.

- [ ] **Step 6: Commit**

```bash
git add Dockerfile fly.toml .gitignore src/rdruid_analyzer/web/cache.py src/rdruid_analyzer/wcl/cache.py
git commit -m "feat(web): add Dockerfile and fly.toml for deployment"
```

---

## Task 10: Deploy to Fly.io

- [ ] **Step 1: Create Fly app and volume**

```bash
fly apps create rdruid-talent-analyzer
fly volumes create data --region ams --size 1
```

- [ ] **Step 2: Set secrets**

```bash
fly secrets set WCL_CLIENT_ID=<your-client-id> WCL_CLIENT_SECRET=<your-client-secret>
```

- [ ] **Step 3: Deploy**

```bash
fly deploy
```

- [ ] **Step 4: Verify**

```bash
fly status
curl https://rdruid-talent-analyzer.fly.dev/api/health
```

Open `https://rdruid-talent-analyzer.fly.dev/` in browser — should show the landing page.

- [ ] **Step 5: End-to-end test on production**

Paste a known report URL, select fight + player, verify results match CLI output.
