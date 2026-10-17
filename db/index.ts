import { env } from "cloudflare:workers";
import { drizzle } from "drizzle-orm/d1";
import * as schema from "./schema";

export function getDb() {
  if (!env.DB) {
    throw new Error(
      "This D1 example requires a Cloudflare Worker with a `DB` binding. The GitHub Pages courses do not use a database."
    );
  }

  return drizzle(env.DB, { schema });
}
