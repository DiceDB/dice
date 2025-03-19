export async function fetchStars() {
  try {
    const response = await fetch("https://api.github.com/repos/DiceDB/dice");
    const data = await response.json();
    const starCount = data.stargazers_count;

    const formattedStars =
      starCount >= 1000
        ? (starCount / 1000).toFixed(1) + "k+"
        : starCount.toString();

    return formattedStars;
  } catch (error) {
    console.error("Failed to fetch GitHub stars:", error);
    return null;
  }
}
