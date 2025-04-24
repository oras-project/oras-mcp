import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";
import { execSync } from "child_process";

export function createServer(): McpServer {
  const server = new McpServer({
    name: "ORAS MCP Server",
    version: "0.1.0",
  });

  server.tool(
    "list_wellknown_registries",
    "List well-known public registries with catalog support.",
    {
      // no parameters
    },
    async () => {
      const registries = [
        {
          name: "mcr.microsoft.com",
          description: "Microsoft Container Registry",
        },
      ];

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(registries, null, 2),
          },
        ],
      };
    },
  );

  server.tool(
    "list_repositories",
    "List repositories of a container registry.",
    {
      registry: z.string().describe("registry name"),
    },
    async ({ registry }) => {
      if (!registry) {
        throw new Error("registry name is required.");
      }

      // call the ORAS CLI to list repositories
      const command = `oras repo list ${registry}`;
      const output = execSync(command, { encoding: "utf-8" });

      // parse the output to get the list of repositories
      const repositories = output.split("\n").filter((line: string) => line.trim() !== "");

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(repositories, null, 2),
          },
        ],
      };
    },
  );

  server.tool(
    "list_tags",
    "List tags in a repository of a container registry.",
    {
      registry: z.string().describe("registry name"),
      repository: z.string().describe("repository name"),
    },
    async ({ registry, repository }) => {
      if (!registry || !repository) {
        throw new Error("registry and repository names are required.");
      }

      // call the ORAS CLI to list tags
      const command = `oras repo tags ${registry}/${repository}`;
      const output = execSync(command, { encoding: "utf-8" });

      // parse the output to get the list of tags
      const tags = output.split("\n").filter((line: string) => line.trim() !== "");

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(tags, null, 2),
          },
        ],
      };
    },
  );

  server.tool(
    "list_referrers",
    "List referrers of a container image or an OCI artifact.",
    {
      registry: z.string().describe("registry name"),
      repository: z.string().describe("repository name"),
      tag: z.string().optional().describe("tag name"),
      digest: z.string().optional().describe("manifest digest"),
    },
    async ({ registry, repository, tag, digest }) => {
      if (!registry || !repository) {
        throw new Error("registry and repository names are required.");
      }
      if (!tag && !digest) {
        throw new Error("Either tag or digest is required.");
      }

      // construct the reference string
      let reference = `${registry}/${repository}`;
      if (tag) {
        reference += `:${tag}`;
      } else if (digest) {
        reference += `@${digest}`;
      }

      // call the ORAS CLI to list referrers
      const command = `oras discover --format json ${reference}`;
      const output = execSync(command, { encoding: "utf-8" });

      return {
        content: [
          {
            type: "text",
            text: output,
          },
        ],
      };
    },
  );

  server.tool(
    "fetch_manifest",
    "Fetch manifest of a container image or an OCI artifact.",
    {
      registry: z.string().describe("registry name"),
      repository: z.string().describe("repository name"),
      tag: z.string().optional().describe("tag name"),
      digest: z.string().optional().describe("manifest digest"),
    },
    async ({ registry, repository, tag, digest }) => {
      if (!registry || !repository) {
        throw new Error("registry and repository names are required.");
      }
      if (!tag && !digest) {
        throw new Error("Either tag or digest is required.");
      }

      // construct the reference string
      let reference = `${registry}/${repository}`;
      if (tag) {
        reference += `:${tag}`;
      } else if (digest) {
        reference += `@${digest}`;
      }

      // call the ORAS CLI to fetch manifest
      const command = `oras manifest fetch --format json ${reference}`;
      const output = execSync(command, { encoding: "utf-8" });

      return {
        content: [
          {
            type: "text",
            text: output,
          },
        ],
      };
    },
  );

  server.tool(
    "fetch_blob",
    "Fetch blob referenced by a digest in a manifest.",
    {
      registry: z.string().describe("registry name"),
      repository: z.string().describe("repository name"),
      digest: z.string().describe("blob digest"),
    },
    async ({ registry, repository, digest }) => {
      if (!registry || !repository || !digest) {
        throw new Error("registry, repository, and digest are required.");
      }

      // construct the reference string
      const reference = `${registry}/${repository}@${digest}`;

      // call the ORAS CLI to fetch blob
      const command = `oras blob fetch -o- ${reference}`;
      const output = execSync(command, { encoding: "utf-8" });

      return {
        content: [
          {
            type: "text",
            text: output,
          },
        ],
      };
    },
  );

  server.tool(
    "parse_reference",
    "Parse a reference string into its components of registry, repository, tag, and digest.",
    {
      reference: z.string().describe("reference string"),
    },
    async ({ reference }) => {
      if (!reference) {
        throw new Error("reference string is required.");
      }

      // parse the reference string
      const components: {
        registry: string;
        repository: string;
        tag?: string;
        digest?: string;
      } = { registry: "", repository: "" };

      const atIndex = reference.indexOf("@");
      if (atIndex !== -1) {
        components.digest = reference.slice(atIndex + 1);
        reference = reference.slice(0, atIndex);
      }

      const colonIndex = reference.indexOf(":");
      if (colonIndex !== -1) {
        components.tag = reference.slice(colonIndex + 1);
        reference = reference.slice(0, colonIndex);
      }
      const slashIndex = reference.indexOf("/");
      if (slashIndex === -1) {
        throw new Error("Invalid reference string format.");
      }
      components.registry = reference.slice(0, slashIndex);
      components.repository = reference.slice(slashIndex + 1);

      if (components.registry === "" || components.repository === "") {
        throw new Error("Invalid reference string format.");
      }

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(components, null, 2),
          },
        ],
      };
    },
  );

  return server;
}
