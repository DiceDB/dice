import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

// https://astro.build/config
export default defineConfig({
  integrations: [
    starlight({
      title: "DiceDB",
      logo: {
        replacesTitle: true,
        light: "./public/dicedb-logo-light.png",
        dark: "./public/dicedb-logo-dark.png",
      },
      customCss: ["./src/styles/docs.css"],
      // useStarlightDarkModeSwitch: false,
      favicon: "/favicon.png",
      editLink: {
        baseUrl: "https://github.com/DiceDB/dice/edit/master/docs/",
      },
      lastUpdated: true,
      expressiveCode: {
        textMarkers: true,
        themes: ["ayu-dark", "light-plus"],
        defaultProps: {
          wrap: true,
        },
        styleOverrides: {
          borderRadius: "0.2rem",
        },
      },
      sidebar: [
        {
          label: "Get started",
          autogenerate: { directory: "get-started" },
        },
        {
          label: "SDK",
          autogenerate: { directory: "sdk" },
        },
        {
          label: "Examples and Tutorials",
          items: [
            {
              label: "Leaderboard",
              link: "https://github.com/DiceDB/dice/tree/master/examples/leaderboard-go",
            },
            {
              label: "CLI Chatroom",
              link: "https://github.com/DiceDB/dice/tree/master/examples/chatroom-go",
            },
          ],
        },
        {
          label: "Commands",
          autogenerate: { directory: "commands" },
        },
      ],
    }),
  ],
});
