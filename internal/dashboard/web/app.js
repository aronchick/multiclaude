// multiclaude Dashboard - Frontend Application
// FORK-ONLY FEATURE

class Dashboard {
    constructor() {
        this.state = null;
        this.eventSource = null;
        this.init();
    }

    async init() {
        // Initial data load
        await this.loadState();

        // Setup SSE for live updates
        this.setupLiveUpdates();

        // Refresh every 30 seconds as fallback
        setInterval(() => this.loadState(), 30000);
    }

    async loadState() {
        try {
            const response = await fetch('/api/state');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }

            this.state = await response.json();
            this.render();
            this.updateTimestamp();
        } catch (err) {
            console.error('Failed to load state:', err);
            this.showError('Failed to load data. Retrying...');
        }
    }

    setupLiveUpdates() {
        this.eventSource = new EventSource('/api/events');

        this.eventSource.onmessage = (event) => {
            try {
                this.state = JSON.parse(event.data);
                this.render();
                this.updateTimestamp();
                this.updateLiveIndicator(true);
            } catch (err) {
                console.error('Failed to parse SSE data:', err);
            }
        };

        this.eventSource.onerror = (err) => {
            console.error('SSE connection error:', err);
            this.updateLiveIndicator(false);

            // Reconnect after 5 seconds
            setTimeout(() => {
                this.eventSource.close();
                this.setupLiveUpdates();
            }, 5000);
        };
    }

    render() {
        if (!this.state) return;

        this.renderOverview();
        this.renderRepositories();
        this.renderHistory();
    }

    renderOverview() {
        let totalRepos = 0;
        let totalAgents = 0;
        let totalWorkers = 0;

        for (const machine of Object.values(this.state.machines)) {
            const repos = Object.values(machine.repos || {});
            totalRepos += repos.length;

            for (const repo of repos) {
                const agents = Object.values(repo.agents || {});
                totalAgents += agents.length;
                totalWorkers += agents.filter(a => a.type === 'worker').length;
            }
        }

        document.getElementById('total-repos').textContent = totalRepos;
        document.getElementById('total-agents').textContent = totalAgents;
        document.getElementById('total-workers').textContent = totalWorkers;
    }

    renderRepositories() {
        const container = document.getElementById('repos-list');

        const repos = [];
        for (const [machineName, machine] of Object.entries(this.state.machines)) {
            for (const [repoName, repo] of Object.entries(machine.repos || {})) {
                repos.push({ name: repoName, machine: machineName, ...repo });
            }
        }

        if (repos.length === 0) {
            container.innerHTML = '<p class="empty">No repositories found</p>';
            return;
        }

        container.innerHTML = repos.map(repo => this.renderRepoCard(repo)).join('');
    }

    renderRepoCard(repo) {
        const agents = Object.entries(repo.agents || {});

        return `
            <div class="repo-card">
                <div class="repo-header">
                    <div>
                        <div class="repo-name">${this.escapeHtml(repo.name)}</div>
                        <div class="repo-url">${this.escapeHtml(repo.github_url || 'Unknown')}</div>
                    </div>
                    <div class="repo-machine">${this.escapeHtml(repo.machine)}</div>
                </div>
                ${agents.length > 0 ? `
                    <div class="agents-grid">
                        ${agents.map(([name, agent]) => this.renderAgentBadge(name, agent)).join('')}
                    </div>
                ` : '<p class="empty">No agents</p>'}
            </div>
        `;
    }

    renderAgentBadge(name, agent) {
        const task = agent.task ? this.escapeHtml(agent.task) : '';
        const taskTitle = task.length > 30 ? task : '';
        const taskDisplay = task.length > 30 ? task.substring(0, 30) + '...' : task;

        return `
            <div class="agent-badge">
                <span class="agent-type agent-type-${agent.type}">${this.escapeHtml(name)}</span>
                ${task ? `<span class="agent-task" title="${taskTitle}">${taskDisplay}</span>` : ''}
            </div>
        `;
    }

    async renderHistory() {
        const container = document.getElementById('history-list');

        // Collect all task history from all repos
        const allHistory = [];
        for (const machine of Object.values(this.state.machines)) {
            for (const [repoName, repo] of Object.entries(machine.repos || {})) {
                const history = repo.task_history || [];
                for (const entry of history) {
                    allHistory.push({ ...entry, repo: repoName });
                }
            }
        }

        // Sort by completion time (most recent first)
        allHistory.sort((a, b) => {
            const timeA = new Date(a.completed_at || a.created_at);
            const timeB = new Date(b.completed_at || b.created_at);
            return timeB - timeA;
        });

        // Limit to 20 most recent
        const recentHistory = allHistory.slice(0, 20);

        if (recentHistory.length === 0) {
            container.innerHTML = '<p class="empty">No task history yet</p>';
            return;
        }

        container.innerHTML = `
            <table class="history-table">
                <thead>
                    <tr>
                        <th>Worker</th>
                        <th>Repository</th>
                        <th>Task</th>
                        <th>Status</th>
                        <th>PR</th>
                        <th>Completed</th>
                    </tr>
                </thead>
                <tbody>
                    ${recentHistory.map(entry => this.renderHistoryRow(entry)).join('')}
                </tbody>
            </table>
        `;
    }

    renderHistoryRow(entry) {
        const completedAt = entry.completed_at
            ? this.formatDate(new Date(entry.completed_at))
            : '-';

        const prLink = entry.pr_url
            ? `<a href="${this.escapeHtml(entry.pr_url)}" target="_blank" class="pr-link">#${entry.pr_number || 'PR'}</a>`
            : '-';

        return `
            <tr>
                <td class="task-name">${this.escapeHtml(entry.name)}</td>
                <td>${this.escapeHtml(entry.repo)}</td>
                <td class="task-desc" title="${this.escapeHtml(entry.task)}">${this.escapeHtml(entry.task)}</td>
                <td><span class="status-badge status-${entry.status}">${this.escapeHtml(entry.status)}</span></td>
                <td>${prLink}</td>
                <td>${completedAt}</td>
            </tr>
        `;
    }

    updateTimestamp() {
        const now = new Date();
        document.getElementById('last-update').textContent =
            `Last update: ${this.formatTime(now)}`;
    }

    updateLiveIndicator(connected) {
        const indicator = document.getElementById('live-indicator');
        if (connected) {
            indicator.style.display = 'flex';
        } else {
            indicator.style.display = 'none';
        }
    }

    formatDate(date) {
        const now = new Date();
        const diff = now - date;
        const minutes = Math.floor(diff / 60000);
        const hours = Math.floor(diff / 3600000);
        const days = Math.floor(diff / 86400000);

        if (minutes < 1) return 'Just now';
        if (minutes < 60) return `${minutes}m ago`;
        if (hours < 24) return `${hours}h ago`;
        if (days < 7) return `${days}d ago`;

        return date.toLocaleDateString();
    }

    formatTime(date) {
        return date.toLocaleTimeString();
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    showError(message) {
        console.error(message);
        // Could add a toast notification here
    }
}

// Initialize dashboard when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => new Dashboard());
} else {
    new Dashboard();
}
