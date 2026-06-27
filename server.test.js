import assert from "node:assert/strict";
import { test } from "node:test";
import { buildScoreboardUrl, normalizeCompetition, toNumber } from "./server.js";

test("buildScoreboardUrl allows valid ESPN query parameters", () => {
  const params = new URLSearchParams({ dates: "20260627", limit: "6" });
  const url = buildScoreboardUrl(params);

  assert.equal(
    url,
    "https://site.api.espn.com/apis/site/v2/sports/soccer/fifa.world/scoreboard?dates=20260627&limit=6"
  );
});

test("buildScoreboardUrl ignores unexpected query parameters", () => {
  const params = new URLSearchParams({ dates: "bad-date", limit: "all", foo: "bar" });
  const url = buildScoreboardUrl(params);

  assert.equal(url, "https://site.api.espn.com/apis/site/v2/sports/soccer/fifa.world/scoreboard");
});

test("toNumber extracts numbers from ESPN display values", () => {
  assert.equal(toNumber("56.7"), 56.7);
  assert.equal(toNumber("1,234"), 1234);
  assert.equal(toNumber("FT"), 0);
});

test("normalizeCompetition produces score, teams, and analytics", () => {
  const event = {
    id: "760475",
    name: "France at Norway",
    shortName: "FRA @ NOR",
    date: "2026-06-27T19:00Z",
    status: {
      displayClock: "74'",
      period: 2,
      type: {
        state: "in",
        description: "In Progress",
        detail: "74'",
        completed: false
      }
    },
    competitions: [
      {
        venue: {
          fullName: "World Cup Stadium",
          address: { city: "Oslo", country: "Norway" }
        },
        competitors: [
          {
            homeAway: "home",
            score: "2",
            team: {
              id: "1",
              displayName: "Norway",
              shortDisplayName: "Norway",
              abbreviation: "NOR",
              color: "ef4444"
            }
          },
          {
            homeAway: "away",
            score: "3",
            team: {
              id: "2",
              displayName: "France",
              shortDisplayName: "France",
              abbreviation: "FRA",
              color: "2563eb"
            }
          }
        ]
      }
    ]
  };
  const summary = {
    boxscore: {
      teams: [
        {
          homeAway: "home",
          statistics: [
            { name: "totalShots", displayValue: "10", label: "Shots" },
            { name: "shotsOnTarget", displayValue: "4", label: "On Goal" },
            { name: "possessionPct", displayValue: "43.3", label: "Possession" },
            { name: "wonCorners", displayValue: "4", label: "Corners" },
            { name: "foulsCommitted", displayValue: "9", label: "Fouls" },
            { name: "yellowCards", displayValue: "1", label: "Yellow Cards" },
            { name: "redCards", displayValue: "0", label: "Red Cards" }
          ]
        },
        {
          homeAway: "away",
          statistics: [
            { name: "totalShots", displayValue: "18", label: "Shots" },
            { name: "shotsOnTarget", displayValue: "9", label: "On Goal" },
            { name: "possessionPct", displayValue: "56.7", label: "Possession" },
            { name: "wonCorners", displayValue: "5", label: "Corners" },
            { name: "foulsCommitted", displayValue: "11", label: "Fouls" },
            { name: "yellowCards", displayValue: "1", label: "Yellow Cards" },
            { name: "redCards", displayValue: "0", label: "Red Cards" }
          ]
        }
      ]
    },
    keyEvents: [{ id: "goal-1", text: "Goal", team: { displayName: "France" } }]
  };

  const game = normalizeCompetition(event, summary);

  assert.equal(game.teams.home.name, "Norway");
  assert.equal(game.teams.away.name, "France");
  assert.deepEqual(game.score, { home: 2, away: 3 });
  assert.equal(game.analytics.totalShots, 28);
  assert.equal(game.analytics.totalOnTarget, 13);
  assert.equal(game.analytics.shotAccuracy, 46);
  assert.equal(game.analytics.goalConversion, 18);
  assert.equal(game.analytics.cards.home, 1);
  assert.equal(game.keyEvents[0].team, "France");
});
