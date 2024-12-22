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

const releases = defineCollection({
  type: "content",
  schema: z.object({
    title: z.string(),
    description: z.string(),
    published_at: z.coerce.date(),
    author: reference("authors"),
  }),
});

const updates = defineCollection({
  type: "content",
  schema: z.object({}),
});

const team = defineCollection({
  type: "data",
  schema: z.object({
    name: z.string(),
    avatar_url: z.string(),
    roles: z.array(z.string()),
    x: z.string(),
    linkedin: z.string(),
    github: z.string(),
    website: z.string(),
  }),
});

export const collections = {
  blog,
  authors,
  releases,
  team,
  updates,
  docs: defineCollection({ schema: docsSchema() }),
};
