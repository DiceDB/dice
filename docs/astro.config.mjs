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
      // themes: ['starlight-theme-light'],
      // useStarlightDarkModeSwitch: false,
      favicon: "/favicon.png",
      sidebar: [
        {
          label: "Get started",
          autogenerate: { directory: "get-started" },
        },
        // {
        // 	label: 'Tutorials',
        // 	autogenerate: { directory: 'tutorials' }
        // },
        {
          label: "SDK",
            autogenerate: { directory: "sdk" },
        },
        {
          label: "Connection Protocols",
          autogenerate: { directory: "protocols" },
        },
        {
          label: "Commands",
          autogenerate: { directory: "commands" },
        },
        {
          label: "Metrics",
          items: [
            {
              label: "Memtier Benchmark",
              link: "https://dicedb.io/metrics/memtier.html",
              attrs: { target: "_blank" },
            },
          ],
        },
      ],
    }),
  ],
});
