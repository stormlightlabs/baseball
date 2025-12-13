(function () {
  "use strict";

  function highlightJSON(data) {
    const json = JSON.stringify(data, null, 2);
    const highlighted = json
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"([^\"]+)":/g, '<span class="json-key">"$1"</span>:')
      .replace(/: "([^\"]*)"/g, ': <span class="json-string">"$1"</span>')
      .replace(/: (\d+\.?\d*)(,?)/g, ': <span class="json-number">$1</span>$2')
      .replace(/: (true|false)(,?)/g, ': <span class="json-boolean">$1</span>$2')
      .replace(/: (null)(,?)/g, ': <span class="json-null">$1</span>$2');
    return `<pre>${highlighted}</pre>`;
  }

  function createChartContext(refs, name) {
    if (!refs) {
      return null;
    }
    const canvas = refs[name];
    if (!canvas) {
      console.warn(`${name} canvas not found`);
      return null;
    }
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      console.warn(`Unable to acquire context for ${name}`);
      return null;
    }
    return ctx;
  }

  function extractArray(value, preferredKeys = []) {
    if (Array.isArray(value)) {
      return value;
    }

    if (!value || typeof value !== "object") {
      return [];
    }

    for (const key of preferredKeys) {
      if (Array.isArray(value[key])) {
        return value[key];
      }
    }

    const defaultKeys = ["leaders", "seasons", "data", "items", "results", "records"];
    for (const key of defaultKeys) {
      if (Array.isArray(value[key])) {
        return value[key];
      }
    }

    for (const val of Object.values(value)) {
      if (Array.isArray(val)) {
        return val;
      }
    }

    return [];
  }

  document.addEventListener("DOMContentLoaded", () => {
    const advancedTab = document.getElementById("advanced-tab");
    if (advancedTab) {
      advancedTab.addEventListener("shown.bs.tab", () => {
        document.dispatchEvent(new CustomEvent("load-advanced-stats"));
      });
    }
  });

  window.basicStatsApp = function basicStatsApp() {
    return {
      pitchingYear: 2023,
      pitchingLeague: "",
      leaderYear: 2023,
      leaderStat: "hr",
      league: "",
      playerId: "troutmi01",
      playerStatType: "batting",
      playerError: "",
      pitchingEndpoint: "",
      leadersEndpoint: "",
      playerEndpoint: "",
      showPitchingJSON: false,
      showLeadersJSON: false,
      showPlayerJSON: false,
      pitchingJSON: "",
      leadersJSON: "",
      playerJSON: "",
      pitchingChart: null,
      leadersChart: null,
      playerChart: null,

      init() {},

      formatJSON(data) {
        return highlightJSON(data);
      },

      getCanvasContext(refName) {
        return createChartContext(this.$refs, refName);
      },

      async loadPitchingLeaders() {
        try {
          let url = `/v1/seasons/${this.pitchingYear}/leaders/pitching?stat=so&limit=10`;
          if (this.pitchingLeague) {
            url += `&league=${this.pitchingLeague}`;
          }

          this.pitchingEndpoint = url;
          const response = await fetch(url);
          const result = await response.json();
          const data = extractArray(result, ["leaders"]);
          this.pitchingJSON = this.formatJSON(data.slice(0, 3));

          if (!data || data.length === 0) {
            console.warn("No pitching leaders data found");
            return;
          }

          const labels = data.map((p) => p.player_id || "Unknown");
          const strikeouts = data.map((p) => p.so || 0);

          if (this.pitchingChart) {
            this.pitchingChart.destroy();
          }

          const ctx = this.getCanvasContext("pitchingChart");
          if (!ctx) {
            return;
          }

          this.pitchingChart = new Chart(ctx, {
            type: "bar",
            data: {
              labels,
              datasets: [
                {
                  label: "Strikeouts",
                  data: strikeouts,
                  backgroundColor: "rgba(54, 162, 235, 0.5)",
                  borderColor: "rgba(54, 162, 235, 1)",
                  borderWidth: 1,
                },
              ],
            },
            options: {
              responsive: true,
              maintainAspectRatio: false,
              indexAxis: "y",
              plugins: {
                title: {
                  display: true,
                  text: `${this.pitchingYear} Strikeout Leaders${
                    this.pitchingLeague ? " - " + this.pitchingLeague : ""
                  }`,
                },
                legend: { display: false },
              },
            },
          });
        } catch (error) {
          console.error("Error loading pitching leaders:", error);
        }
      },

      async loadBattingLeaders() {
        try {
          let url = `/v1/seasons/${this.leaderYear}/leaders/batting?stat=${this.leaderStat}&limit=10`;
          if (this.league) {
            url += `&league=${this.league}`;
          }

          this.leadersEndpoint = url;
          const response = await fetch(url);
          const result = await response.json();
          const data = extractArray(result, ["leaders"]);
          this.leadersJSON = this.formatJSON(data.slice(0, 3));

          if (!data || data.length === 0) {
            console.warn("No leaders data found");
            return;
          }

          const labels = data.map((p) => p.player_id || "Unknown");
          const values = data.map((p) => {
            const val = p[this.leaderStat];
            return val !== undefined ? val : 0;
          });

          if (this.leadersChart) {
            this.leadersChart.destroy();
          }

          const ctx = this.getCanvasContext("leadersChart");
          if (!ctx) {
            return;
          }

          this.leadersChart = new Chart(ctx, {
            type: "bar",
            data: {
              labels,
              datasets: [
                {
                  label: this.leaderStat.toUpperCase(),
                  data: values,
                  backgroundColor: "rgba(54, 162, 235, 0.5)",
                  borderColor: "rgba(54, 162, 235, 1)",
                  borderWidth: 1,
                },
              ],
            },
            options: {
              responsive: true,
              maintainAspectRatio: false,
              indexAxis: "y",
              plugins: {
                title: {
                  display: true,
                  text: `${this.leaderYear} ${this.leaderStat.toUpperCase()} Leaders${
                    this.league ? " - " + this.league : ""
                  }`,
                },
                legend: { display: false },
              },
            },
          });
        } catch (error) {
          console.error("Error loading leaders:", error);
        }
      },

      async loadPlayerCareer() {
        this.playerError = "";
        try {
          const endpoint = this.playerStatType === "batting" ? "batting" : "pitching";
          const url = `/v1/players/${this.playerId}/stats/${endpoint}`;
          this.playerEndpoint = url;
          const response = await fetch(url);

          if (!response.ok) {
            this.playerError = `Player not found: ${this.playerId}`;
            this.playerEndpoint = "";
            return;
          }

          const payload = await response.json();
          const data = extractArray(payload, ["seasons"]);
          this.playerJSON = this.formatJSON(data.slice(0, 3));

          if (!data || data.length === 0) {
            this.playerError = "No career data found for this player";
            return;
          }

          const labels = data.map((s) => s.year);

          if (this.playerChart) {
            this.playerChart.destroy();
          }

          const ctx = this.getCanvasContext("playerChart");
          if (!ctx) {
            return;
          }

          if (this.playerStatType === "batting") {
            this.playerChart = new Chart(ctx, {
              type: "line",
              data: {
                labels,
                datasets: [
                  {
                    label: "AVG",
                    data: data.map((s) => s.avg),
                    borderColor: "rgb(255, 99, 132)",
                    yAxisID: "y",
                    tension: 0.1,
                  },
                  {
                    label: "OBP",
                    data: data.map((s) => s.obp),
                    borderColor: "rgb(54, 162, 235)",
                    yAxisID: "y",
                    tension: 0.1,
                  },
                  {
                    label: "SLG",
                    data: data.map((s) => s.slg),
                    borderColor: "rgb(75, 192, 192)",
                    yAxisID: "y",
                    tension: 0.1,
                  },
                  {
                    label: "HR",
                    data: data.map((s) => s.hr),
                    borderColor: "rgb(255, 206, 86)",
                    yAxisID: "y1",
                    tension: 0.1,
                  },
                ],
              },
              options: {
                responsive: true,
                maintainAspectRatio: false,
                interaction: { mode: "index", intersect: false },
                plugins: {
                  title: { display: true, text: `${this.playerId} Career Batting Stats` },
                },
                scales: {
                  y: {
                    type: "linear",
                    display: true,
                    position: "left",
                    title: { display: true, text: "AVG/OBP/SLG" },
                  },
                  y1: {
                    type: "linear",
                    display: true,
                    position: "right",
                    title: { display: true, text: "HR" },
                    grid: { drawOnChartArea: false },
                  },
                },
              },
            });
          } else {
            this.playerChart = new Chart(ctx, {
              type: "line",
              data: {
                labels,
                datasets: [
                  {
                    label: "ERA",
                    data: data.map((s) => s.era),
                    borderColor: "rgb(255, 99, 132)",
                    yAxisID: "y",
                    tension: 0.1,
                  },
                  {
                    label: "WHIP",
                    data: data.map((s) => s.whip),
                    borderColor: "rgb(54, 162, 235)",
                    yAxisID: "y",
                    tension: 0.1,
                  },
                  {
                    label: "K/9",
                    data: data.map((s) => s.k_per_9),
                    borderColor: "rgb(75, 192, 192)",
                    yAxisID: "y1",
                    tension: 0.1,
                  },
                ],
              },
              options: {
                responsive: true,
                maintainAspectRatio: false,
                interaction: { mode: "index", intersect: false },
                plugins: {
                  title: { display: true, text: `${this.playerId} Career Pitching Stats` },
                },
                scales: {
                  y: {
                    type: "linear",
                    display: true,
                    position: "left",
                    title: { display: true, text: "ERA/WHIP" },
                  },
                  y1: {
                    type: "linear",
                    display: true,
                    position: "right",
                    title: { display: true, text: "K/9" },
                    grid: { drawOnChartArea: false },
                  },
                },
              },
            });
          }
        } catch (error) {
          console.error("Error loading player career:", error);
          this.playerError = "Error loading player data";
        }
      },
    };
  };

  window.advancedStatsApp = function advancedStatsApp() {
    return {
      playerIdAdv: "troutmi01",
      advSeasonPlayer: 2023,
      advancedYear: 2023,
      minPA: 502,
      playerIdWAR: "troutmi01",
      warSeason: 2023,
      playerAdvData: null,
      warData: null,
      playerAdvError: "",
      warError: "",
      playerAdvEndpoint: "",
      advLeadersEndpoint: "",
      warEndpoint: "",
      showPlayerAdvJSON: false,
      showAdvLeadersJSON: false,
      showWARJSON: false,
      playerAdvJSON: "",
      advLeadersJSON: "",
      warJSON: "",
      advancedLeadersChart: null,

      init() {
        let loaded = false;
        document.addEventListener("load-advanced-stats", () => {
          if (!loaded) {
            loaded = true;
            this.$nextTick(() => this.loadAdvancedStats());
          }
        });

        const advTab = document.getElementById("advanced-tab");
        if (advTab && advTab.classList.contains("active")) {
          loaded = true;
          this.$nextTick(() => this.loadAdvancedStats());
        }
      },

      formatJSON(data) {
        return highlightJSON(data);
      },

      getCanvasContext(refName) {
        return createChartContext(this.$refs, refName);
      },

      async loadPlayerAdvancedBatting() {
        this.playerAdvError = "";
        this.playerAdvData = null;
        try {
          const url = `/v1/players/${this.playerIdAdv}/stats/batting/advanced?season=${this.advSeasonPlayer}`;
          this.playerAdvEndpoint = url;
          const response = await fetch(url);

          if (!response.ok) {
            this.playerAdvError = `Failed to load advanced stats for ${this.playerIdAdv}`;
            this.playerAdvEndpoint = "";
            return;
          }

          const data = await response.json();
          this.playerAdvData = data;
          this.playerAdvJSON = this.formatJSON(data);
        } catch (error) {
          console.error("Error loading advanced batting:", error);
          this.playerAdvError = "Error loading advanced batting stats";
        }
      },

      async loadAdvancedStats() {
        try {
          const url = `/v1/stats/batting?season=${this.advancedYear}&min_ab=${this.minPA}&sort_by=woba&sort_order=desc&per_page=10`;
          this.advLeadersEndpoint = url;
          const response = await fetch(url);
          const result = await response.json();
          const data = extractArray(result, ["data"]);
          this.advLeadersJSON = this.formatJSON(data.slice(0, 3));

          if (!data || data.length === 0) {
            console.warn("No advanced stats data found");
            return;
          }

          const labels = data.map((p) => p.player_id || "Unknown");
          const woba = data.map((p) => p.woba || p.obp || 0);
          const wrcPlus = data.map((p) => p.wrc_plus || p.ops * 100 || 0);

          if (this.advancedLeadersChart) {
            this.advancedLeadersChart.destroy();
          }

          const ctx = this.getCanvasContext("advancedLeadersChart");
          if (!ctx) {
            return;
          }

          this.advancedLeadersChart = new Chart(ctx, {
            type: "bar",
            data: {
              labels,
              datasets: [
                {
                  label: "wOBA (or OBP)",
                  data: woba,
                  backgroundColor: "rgba(153, 102, 255, 0.5)",
                  borderColor: "rgba(153, 102, 255, 1)",
                  borderWidth: 1,
                  yAxisID: "y",
                },
                {
                  label: "wRC+ (or OPS*100)",
                  data: wrcPlus,
                  backgroundColor: "rgba(255, 159, 64, 0.5)",
                  borderColor: "rgba(255, 159, 64, 1)",
                  borderWidth: 1,
                  yAxisID: "y1",
                },
              ],
            },
            options: {
              responsive: true,
              maintainAspectRatio: false,
              plugins: {
                title: {
                  display: true,
                  text: `${this.advancedYear} Advanced Stats (Min ${this.minPA} AB)`,
                },
              },
              scales: {
                y: {
                  type: "linear",
                  display: true,
                  position: "left",
                  title: { display: true, text: "wOBA" },
                },
                y1: {
                  type: "linear",
                  display: true,
                  position: "right",
                  title: { display: true, text: "wRC+" },
                  grid: { drawOnChartArea: false },
                },
              },
            },
          });
        } catch (error) {
          console.error("Error loading advanced stats:", error);
        }
      },

      async loadPlayerWAR() {
        this.warError = "";
        this.warData = null;
        try {
          const url = `/v1/players/${this.playerIdWAR}/stats/war?season=${this.warSeason}`;
          this.warEndpoint = url;
          const response = await fetch(url);

          if (!response.ok) {
            this.warError = `Failed to load WAR for ${this.playerIdWAR}`;
            this.warEndpoint = "";
            return;
          }

          const data = await response.json();
          this.warData = data;
          this.warJSON = this.formatJSON(data);
        } catch (error) {
          console.error("Error loading WAR:", error);
          this.warError = "Error loading WAR stats";
        }
      },
    };
  };
})();
