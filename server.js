import { createServer } from "node:http";
import { readFile } from "node:fs/promises";
import { extname, join, normalize } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const __dirname = fileURLToPath(new URL(".", import.meta.url));
const publicDir = join(__dirname, "public");
const ESPN_BASE = "https://site.api.espn.com/apis/site/v2/sports/soccer/fifa.world";
const CACHE_MS = 15_000;
const PORT = Number(process.env.PORT || 3000);

const cache = new Map();

const mimeTypes = {
  ".html": "text/html; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".js": "application/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".svg": "image/svg+xml",
  ".png": "image/png",
  ".ico": "image/x-icon"
};

function sendJson(res, statusCode, payload) {
  res.writeHead(statusCode, {
    "content-type": "application/json; charset=utf-8",
    "cache-control": "no-store"
  });
  res.end(JSON.stringify(payload));
}

async function fetchJson(url) {
  const cached = cache.get(url);
  if (cached && Date.now() - cached.createdAt < CACHE_MS) {
    return cached.value;
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 12_000);

  try {
    const response = await fetch(url, {
      signal: controller.signal,
      headers: {
        "accept": "application/json",
        "user-agent": "world-cup-live-analytics/1.0"
      }
    });

    if (!response.ok) {
      throw new Error(`ESPN responded ${response.status} for ${url}`);
    }

    const value = await response.json();
    cache.set(url, { createdAt: Date.now(), value });
    return value;
  } finally {
    clearTimeout(timeout);
  }
}

function buildScoreboardUrl(searchParams) {
  const url = new URL(`${ESPN_BASE}/scoreboard`);
  const dates = searchParams.get("dates");
  const limit = searchParams.get("limit");

  if (dates && /^\d{8}$/.test(dates)) {
    url.searchParams.set("dates", dates);
  }

  if (limit && /^\d{1,2}$/.test(limit)) {
    url.searchParams.set("limit", limit);
  }

  return url.toString();
}

function toNumber(value, fallback = 0) {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === "string") {
    const match = value.replace(/,/g, "").match(/-?\d+(\.\d+)?/);
    if (match) {
      return Number(match[0]);
    }
  }

  return fallback;
}

function statMap(teamStats = []) {
  return Object.fromEntries(
    teamStats.map((stat) => [
      stat.name,
      {
        label: stat.label || stat.displayName || stat.name,
        value: toNumber(stat.value ?? stat.displayValue),
        displayValue: stat.displayValue ?? String(stat.value ?? "")
      }
    ])
  );
}

function getStat(stats, name) {
  return stats?.[name]?.value ?? 0;
}

function analyticsFor(homeStats, awayStats, homeScore, awayScore) {
  const homeShots = getStat(homeStats, "totalShots");
  const awayShots = getStat(awayStats, "totalShots");
  const homeOnTarget = getStat(homeStats, "shotsOnTarget");
  const awayOnTarget = getStat(awayStats, "shotsOnTarget");
  const totalShots = homeShots + awayShots;
  const totalOnTarget = homeOnTarget + awayOnTarget;
  const homePossession = getStat(homeStats, "possessionPct");
  const awayPossession = getStat(awayStats, "possessionPct");
  const homePressure = Math.round(
    homeScore * 18 +
      homeOnTarget * 6 +
      homeShots * 2 +
      getStat(homeStats, "wonCorners") * 2 +
      homePossession * 0.18 -
      getStat(homeStats, "foulsCommitted")
  );
  const awayPressure = Math.round(
    awayScore * 18 +
      awayOnTarget * 6 +
      awayShots * 2 +
      getStat(awayStats, "wonCorners") * 2 +
      awayPossession * 0.18 -
      getStat(awayStats, "foulsCommitted")
  );

  return {
    totalShots,
    totalOnTarget,
    shotAccuracy: totalShots ? Math.round((totalOnTarget / totalShots) * 100) : null,
    goalConversion: totalShots ? Math.round(((homeScore + awayScore) / totalShots) * 100) : null,
    possessionSwing: Math.round((homePossession - awayPossession) * 10) / 10,
    pressure: {
      home: Math.max(homePressure, 0),
      away: Math.max(awayPressure, 0)
    },
    cards: {
      home: getStat(homeStats, "yellowCards") + getStat(homeStats, "redCards"),
      away: getStat(awayStats, "yellowCards") + getStat(awayStats, "redCards")
    }
  };
}

function normalizeCompetition(event, summary) {
  const competition = event.competitions?.[0] || {};
  const competitors = competition.competitors || [];
  const home = competitors.find((team) => team.homeAway === "home") || competitors[0] || {};
  const away = competitors.find((team) => team.homeAway === "away") || competitors[1] || {};
  const summaryTeams = summary?.boxscore?.teams || [];
  const statsByHomeAway = Object.fromEntries(
    summaryTeams.map((team) => [team.homeAway, statMap(team.statistics || [])])
  );
  const homeStats = statsByHomeAway.home || {};
  const awayStats = statsByHomeAway.away || {};
  const homeScore = toNumber(home.score);
  const awayScore = toNumber(away.score);

  return {
    id: event.id,
    name: event.name,
    shortName: event.shortName,
    date: event.date,
    venue: competition.venue?.fullName || "Venue TBD",
    city: [competition.venue?.address?.city, competition.venue?.address?.country]
      .filter(Boolean)
      .join(", "),
    status: {
      state: event.status?.type?.state,
      description: event.status?.type?.description,
      detail: event.status?.type?.detail || event.status?.type?.shortDetail,
      completed: Boolean(event.status?.type?.completed),
      clock: event.status?.displayClock || "",
      period: event.status?.period || 0
    },
    teams: {
      home: normalizeTeam(home, homeStats),
      away: normalizeTeam(away, awayStats)
    },
    score: {
      home: homeScore,
      away: awayScore
    },
    analytics: analyticsFor(homeStats, awayStats, homeScore, awayScore),
    keyEvents: (summary?.keyEvents || []).slice(0, 8).map((item) => ({
      id: item.id,
      clock: item.clock?.displayValue || item.displayTime || "",
      team: item.team?.displayName || item.team?.shortDisplayName || "",
      type: item.type?.text || item.type?.displayName || item.text || "",
      text: item.text || item.shortText || ""
    }))
  };
}

function normalizeTeam(competitor, stats) {
  const team = competitor.team || {};
  return {
    id: team.id,
    name: team.displayName || team.name || "TBD",
    shortName: team.shortDisplayName || team.abbreviation || team.name || "TBD",
    abbreviation: team.abbreviation || "",
    color: team.color ? `#${team.color.replace(/^#/, "")}` : "#38bdf8",
    logo: team.logo || team.logos?.[0]?.href || "",
    score: toNumber(competitor.score),
    winner: Boolean(competitor.winner),
    stats
  };
}

async function getDashboardData(searchParams = new URLSearchParams()) {
  const scoreboard = await fetchJson(buildScoreboardUrl(searchParams));
  const events = scoreboard.events || [];
  const summaries = await Promise.all(
    events.map(async (event) => {
      try {
        return await fetchJson(`${ESPN_BASE}/summary?event=${event.id}`);
      } catch (error) {
        return { error: error.message };
      }
    })
  );

  const games = events.map((event, index) => normalizeCompetition(event, summaries[index]));
  const liveGames = games.filter((game) => game.status.state === "in").length;
  const completedGames = games.filter((game) => game.status.completed).length;
  const totalGoals = games.reduce((sum, game) => sum + game.score.home + game.score.away, 0);

  return {
    source: "ESPN public FIFA World Cup API",
    fetchedAt: new Date().toISOString(),
    refreshSeconds: Math.round(CACHE_MS / 1000),
    summary: {
      games: games.length,
      liveGames,
      completedGames,
      totalGoals,
      averageGoals: games.length ? Math.round((totalGoals / games.length) * 10) / 10 : 0
    },
    games
  };
}

async function serveStatic(req, res) {
  const url = new URL(req.url, "http://localhost");
  const requestPath = url.pathname === "/" ? "/index.html" : decodeURIComponent(url.pathname);
  const normalizedPath = normalize(requestPath).replace(/^(\.\.[/\\])+/, "");
  const filePath = join(publicDir, normalizedPath);

  if (!filePath.startsWith(publicDir)) {
    res.writeHead(403);
    res.end("Forbidden");
    return;
  }

  try {
    const body = await readFile(filePath);
    res.writeHead(200, {
      "content-type": mimeTypes[extname(filePath)] || "application/octet-stream",
      "cache-control": "no-cache"
    });
    res.end(body);
  } catch {
    res.writeHead(404, { "content-type": "text/plain; charset=utf-8" });
    res.end("Not found");
  }
}

export function createApp() {
  return createServer(async (req, res) => {
    try {
      const url = new URL(req.url, "http://localhost");

      if (url.pathname === "/api/scoreboard") {
        sendJson(res, 200, await getDashboardData(url.searchParams));
        return;
      }

      await serveStatic(req, res);
    } catch (error) {
      sendJson(res, 502, {
        error: "Unable to load live World Cup data",
        detail: error.message
      });
    }
  });
}

export {
  analyticsFor,
  buildScoreboardUrl,
  getDashboardData,
  normalizeCompetition,
  statMap,
  toNumber
};

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  createApp().listen(PORT, () => {
    console.log(`World Cup live analytics running at http://localhost:${PORT}`);
  });
}
