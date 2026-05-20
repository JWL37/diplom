import createClient from "openapi-fetch";
import type { components, paths } from "./schema";

export const apiClient = createClient<paths>({
  baseUrl: ""
});

export type ClassListItem = components["schemas"]["ClassListItem"];
export type Class = components["schemas"]["Class"];
export type Golden = components["schemas"]["Golden"];
export type GoldenSearchResult = components["schemas"]["GoldenSearchResult"];
