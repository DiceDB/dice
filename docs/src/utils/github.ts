export async function fetchRepoDetails() {
  try {
    const response = await fetch("https://api.github.com/repos/DiceDB/dice");
    if (!response.ok) {
      throw new Error(
        `GitHub API request failed with status ${response.status}`,
      );
    }

    const data = await response.json();

    const formattedStars =
      data?.stargazers_count >= 1000
        ? (data?.stargazers_count / 1000).toFixed(1) + "k+"
        : data?.stargazers_count.toString();

    return {
      name: data.name,
      stars: formattedStars,
      forks: data.forks_count,
      watchers: data.watchers_count,
      openIssues: data.open_issues_count,
      url: data.html_url,
    };
  } catch (error) {
    console.error("🚨 GitHub API Error:", error);
    throw new Error("❌ Failed to fetch GitHub repository data. Build halted.");
  }
}
