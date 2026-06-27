const state = {
  data: null,
  selectedId: null,
  timer: null,
  loading: false
};

const elements = {
  averageGoals: document.querySelector("#averageGoals"),
  connectionState: document.querySelector("#connectionState"),
  gamesTracked: document.querySelector("#gamesTracked"),
  lastUpdated: document.querySelector("#lastUpdated"),
  liveNow: document.querySelector("#liveNow"),
  matchDetail: document.querySelector("#matchDetail"),
  matchList: document.querySelector("#matchList"),
  refreshButton: document.querySelector("#refreshButton"),
  sourceLine: document.querySelector("#sourceLine"),
  totalGoals: document.querySelector("#totalGoals")
};

function escapeHtml(value = "") {
  return String(value).replace(/[&<>"']/g, (char) => {
    const entities = {
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#039;"
    };
    return entities[char];
  });
}

function formatDate(value) {
  return new Intl.DateTimeFormat(undefined, {
    weekday: "short",
    hour: "numeric",
    minute: "2-digit",
    month: "short",
    day: "numeric"
  }).format(new Date(value));
}

function formatUpdated(value) {
  return new Intl.DateTimeFormat(undefined, {
    hour: "numeric",
    minute: "2-digit",
    second: "2-digit"
  }).format(new Date(value));
}

function teamLogo(team) {
  if (!team.logo) {
    return `<span aria-hidden="true">${escapeHtml(team.abbreviation || team.shortName.slice(0, 3))}</span>`;
  }

  return `<img src="${escapeHtml(team.logo)}" alt="" loading="lazy">`;
}

function stat(team, name, fallback = "0") {
  const value = team.stats?.[name]?.displayValue ?? team.stats?.[name]?.value;
  return value === undefined || value === "" ? fallback : value;
}

function statNumber(team, name) {
  return Number(team.stats?.[name]?.value || 0);
}

function splitPercent(left, right) {
  const total = left + right;
  return total > 0 ? Math.round((left / total) * 100) : 50;
}

function statusClass(game) {
  if (game.status.state === "in") {
    return "is-live";
  }

  if (game.status.completed) {
    return "";
  }

  return "is-upcoming";
}

function statusLabel(game) {
  if (game.status.state === "in") {
    return game.status.clock ? `Live ${game.status.clock}` : "Live";
  }

  return game.status.description || "Scheduled";
}

function setConnection(status, message) {
  elements.connectionState.textContent = message;
  elements.connectionState.classList.toggle("is-live", status === "ok");
  elements.connectionState.classList.toggle("is-error", status === "error");
}

function renderSummary(data) {
  elements.gamesTracked.textContent = data.summary.games;
  elements.liveNow.textContent = data.summary.liveGames;
  elements.totalGoals.textContent = data.summary.totalGoals;
  elements.averageGoals.textContent = data.summary.averageGoals.toFixed(1);
  elements.lastUpdated.textContent = formatUpdated(data.fetchedAt);
  elements.sourceLine.textContent = `Source: ${data.source} - refreshed ${formatUpdated(data.fetchedAt)}`;
}

function renderMatchList(games) {
  if (!games.length) {
    elements.matchList.innerHTML = `
      <div class="empty-state">
        <h3>No World Cup games found</h3>
        <p>Try refreshing when fixtures are available.</p>
      </div>
    `;
    return;
  }

  elements.matchList.innerHTML = games
    .map((game) => {
      const selected = game.id === state.selectedId ? "is-selected" : "";
      return `
        <button class="match-card ${selected}" type="button" data-match-id="${escapeHtml(game.id)}">
          <div class="match-card-header">
            <span class="match-status ${statusClass(game)}">${escapeHtml(statusLabel(game))}</span>
            <time datetime="${escapeHtml(game.date)}">${escapeHtml(formatDate(game.date))}</time>
          </div>
          ${renderCompactTeam(game.teams.away, game.score.away)}
          ${renderCompactTeam(game.teams.home, game.score.home)}
        </button>
      `;
    })
    .join("");

  elements.matchList.querySelectorAll("[data-match-id]").forEach((button) => {
    button.addEventListener("click", () => {
      state.selectedId = button.dataset.matchId;
      render();
    });
  });
}

function renderCompactTeam(team, score) {
  return `
    <div class="match-row">
      <span class="team-name">${teamLogo(team)} ${escapeHtml(team.shortName)}</span>
      <span class="score">${score}</span>
    </div>
  `;
}

function render() {
  const data = state.data;
  if (!data) {
    return;
  }

  if (!state.selectedId || !data.games.some((game) => game.id === state.selectedId)) {
    state.selectedId = data.games.find((game) => game.status.state === "in")?.id || data.games[0]?.id || null;
  }

  renderSummary(data);
  renderMatchList(data.games);

  const selected = data.games.find((game) => game.id === state.selectedId);
  elements.matchDetail.innerHTML = selected ? renderMatchDetail(selected) : renderEmptyDetail();
}

function renderMatchDetail(game) {
  const { home, away } = game.teams;
  const analytics = game.analytics;

  return `
    <div class="detail-topline">
      <div>
        <p class="eyebrow">Selected game</p>
        <h2>${escapeHtml(game.shortName || game.name)}</h2>
        <p class="detail-meta">${escapeHtml(game.venue)} ${game.city ? `- ${escapeHtml(game.city)}` : ""}</p>
      </div>
      <span class="match-status ${statusClass(game)}">${escapeHtml(statusLabel(game))}</span>
    </div>

    <div class="detail-scoreboard">
      ${renderDetailTeam(away)}
      <div class="big-score">${game.score.away} - ${game.score.home}</div>
      ${renderDetailTeam(home)}
    </div>

    <div class="analytics-grid">
      ${metricCard("Total shots", analytics.totalShots || "0", "Combined attempts")}
      ${metricCard("On target", analytics.totalOnTarget || "0", `${analytics.shotAccuracy ?? 0}% accuracy`)}
      ${metricCard("Goal conversion", `${analytics.goalConversion ?? 0}%`, "Goals per shot")}
    </div>

    <section class="comparison-panel">
      <h3>Live team comparison</h3>
      ${barRow("Pressure index", analytics.pressure.away, analytics.pressure.home, away.shortName, home.shortName)}
      ${barRow("Possession", statNumber(away, "possessionPct"), statNumber(home, "possessionPct"), away.shortName, home.shortName, "%")}
      ${barRow("Shots", statNumber(away, "totalShots"), statNumber(home, "totalShots"), away.shortName, home.shortName)}
      ${barRow("Corners", statNumber(away, "wonCorners"), statNumber(home, "wonCorners"), away.shortName, home.shortName)}
      ${barRow("Cards", analytics.cards.away, analytics.cards.home, away.shortName, home.shortName)}
    </section>

    <section class="timeline">
      <h3>Key match events</h3>
      ${renderEvents(game.keyEvents)}
    </section>
  `;
}

function renderDetailTeam(team) {
  return `
    <div class="detail-team">
      ${teamLogo(team)}
      <strong>${escapeHtml(team.name)}</strong>
      <span>${escapeHtml(stat(team, "shotsOnTarget"))} on target - ${escapeHtml(stat(team, "possessionPct"))}% possession</span>
    </div>
  `;
}

function metricCard(label, value, caption) {
  return `
    <article class="metric-card">
      <span>${escapeHtml(label)}</span>
      <strong>${escapeHtml(value)}</strong>
      <span>${escapeHtml(caption)}</span>
    </article>
  `;
}

function barRow(label, awayValue, homeValue, awayName, homeName, suffix = "") {
  const awayPercent = splitPercent(awayValue, homeValue);
  const homePercent = 100 - awayPercent;

  return `
    <div class="bar-row">
      <div class="bar-label">
        <span>${escapeHtml(awayName)} ${awayValue}${suffix}</span>
        <strong>${escapeHtml(label)}</strong>
        <span>${homeValue}${suffix} ${escapeHtml(homeName)}</span>
      </div>
      <div class="split-bar">
        <span style="width: ${awayPercent}%"></span>
        <span style="width: ${homePercent}%"></span>
      </div>
    </div>
  `;
}

function renderEvents(events) {
  if (!events.length) {
    return "<p class=\"detail-meta\">No key events available yet.</p>";
  }

  return `
    <ul class="event-list">
      ${events
        .map(
          (event) => `
            <li>
              <strong>${escapeHtml(event.clock || "--")}</strong>
              <span>${escapeHtml(event.team ? `${event.team}: ` : "")}${escapeHtml(event.text || event.type)}</span>
            </li>
          `
        )
        .join("")}
    </ul>
  `;
}

function renderEmptyDetail() {
  return `
    <div class="empty-state">
      <h2>No match selected</h2>
      <p>Select a World Cup game to view analytics.</p>
    </div>
  `;
}

function renderError(error) {
  elements.matchDetail.innerHTML = `
    <div class="error-state">
      <h2>Live feed unavailable</h2>
      <p>${escapeHtml(error.message || "The data source could not be reached.")}</p>
    </div>
  `;
}

async function loadDashboard() {
  if (state.loading) {
    return;
  }

  state.loading = true;
  elements.refreshButton.disabled = true;
  setConnection("loading", "Updating");

  try {
    const response = await fetch("/api/scoreboard", { cache: "no-store" });
    const payload = await response.json();

    if (!response.ok) {
      throw new Error(payload.detail || payload.error || "Unable to load scoreboard");
    }

    state.data = payload;
    setConnection("ok", payload.summary.liveGames ? "Live data" : "Connected");
    render();
    scheduleRefresh(payload.refreshSeconds);
  } catch (error) {
    setConnection("error", "Feed error");
    renderError(error);
  } finally {
    state.loading = false;
    elements.refreshButton.disabled = false;
  }
}

function scheduleRefresh(seconds = 15) {
  clearTimeout(state.timer);
  state.timer = setTimeout(loadDashboard, seconds * 1000);
}

elements.refreshButton.addEventListener("click", loadDashboard);
loadDashboard();
