import { defineCollection, z, reference } from "astro:content";
import { docsSchema } from "@astrojs/starlight/schema";

const blog = defineCollection({
  type: "content",
  schema: z.object({
    title: z.string(),
    description: z.string(),
    published_at: z.coerce.date(),
    author: reference("authors"),
  }),
});

const authors = defineCollection({
  type: "data",
  schema: z.object({
    name: z.string(),
    bio: z.string(),
    avatar_url: z.string(),
    url: z.string(),
  }),
});

export const collections = {
  blog,
  authors,
  docs: defineCollection({ schema: docsSchema() }),
};
