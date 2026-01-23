package fork

import (
	"testing"
)

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantOwner   string
		wantRepo    string
		wantErr     bool
	}{
		{
			name:      "HTTPS with .git",
			url:       "https://github.com/owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS without .git",
			url:       "https://github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "SSH with .git",
			url:       "git@github.com:owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "SSH without .git",
			url:       "git@github.com:owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS with complex owner",
			url:       "https://github.com/my-org/my-repo",
			wantOwner: "my-org",
			wantRepo:  "my-repo",
			wantErr:   false,
		},
		{
			name:      "SSH with underscores",
			url:       "git@github.com:user_name/repo_name.git",
			wantOwner: "user_name",
			wantRepo:  "repo_name",
			wantErr:   false,
		},
		{
			name:    "Invalid URL",
			url:     "not-a-github-url",
			wantErr: true,
		},
		{
			name:    "GitLab URL",
			url:     "https://gitlab.com/owner/repo",
			wantErr: true,
		},
		{
			name:    "Missing repo",
			url:     "https://github.com/owner",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseGitHubURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGitHubURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("ParseGitHubURL() owner = %v, want %v", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("ParseGitHubURL() repo = %v, want %v", repo, tt.wantRepo)
				}
			}
		})
	}
}

func TestForkInfo(t *testing.T) {
	// Test ForkInfo struct defaults
	info := &ForkInfo{
		IsFork:      true,
		OriginURL:   "https://github.com/me/repo",
		OriginOwner: "me",
		OriginRepo:  "repo",
		UpstreamURL: "https://github.com/upstream/repo",
		UpstreamOwner: "upstream",
		UpstreamRepo:  "repo",
	}

	if !info.IsFork {
		t.Error("Expected IsFork to be true")
	}
	if info.OriginOwner != "me" {
		t.Errorf("Expected OriginOwner to be 'me', got %s", info.OriginOwner)
	}
	if info.UpstreamOwner != "upstream" {
		t.Errorf("Expected UpstreamOwner to be 'upstream', got %s", info.UpstreamOwner)
	}
}
