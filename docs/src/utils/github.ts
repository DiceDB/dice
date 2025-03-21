interface GitHubRepository {
  stargazers_count: number;
  name: string;
  description: string;
  html_url: string;
  language: string;
  forks_count: number;
  updated_at: string;
}

export async function fetchGitHubRepository(): Promise<GitHubRepository> {
  const response = await fetch('https://api.github.com/repos/DiceDB/dice');
  if (!response.ok) {
    throw new Error(`GitHub API error: ${response.status}`);
  }
  return response.json();
}

export function formatRepositoryStats(repository: GitHubRepository): string {
  const { stargazers_count } = repository;
  if (stargazers_count >= 1000) {
    const k = (stargazers_count / 1000).toFixed(1);
    return `${k}k+`;
  }
  return stargazers_count.toString();
}