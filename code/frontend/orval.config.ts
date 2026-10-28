import { defineConfig } from "orval";

export default defineConfig({
  api: {
    input: {
      target: "./swagger.json",
    },
    output: {
      mode: "tags-split",
      target: "./src/generated/api",
      schemas: "./src/generated/model",
      client: "react-query",
      mock: false,
      clean: true,
      override: {
        mutator: {
          path: "./src/lib/orval/mutator.ts",
          name: "orvalAxios",
        },
      },
    },
  },
});
