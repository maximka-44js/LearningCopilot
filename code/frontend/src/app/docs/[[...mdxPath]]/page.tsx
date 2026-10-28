import { importPage } from "nextra/pages";
import { notFound } from "next/navigation";

function toContentPathSegments(mdxPath?: string[]): string[] {
  if (!mdxPath || mdxPath.length === 0) {
    return ["docs"];
  }

  return ["docs", ...mdxPath];
}

type DocsPageProps = {
  params: Promise<{
    mdxPath?: string[];
  }>;
};

export async function generateMetadata(props: DocsPageProps) {
  const params = await props.params;

  try {
    const { metadata } = await importPage(toContentPathSegments(params.mdxPath));
    return metadata;
  } catch {
    return {
      title: "Documentation",
      description: "Project documentation",
    };
  }
}

export default async function DocsPage(props: DocsPageProps) {
  const params = await props.params;
  let result: Awaited<ReturnType<typeof importPage>>;

  try {
    result = await importPage(toContentPathSegments(params.mdxPath));
  } catch {
    notFound();
  }

  const { default: MDXContent, toc, metadata } = result;
  return <MDXContent params={params} toc={toc} metadata={metadata} />;
}
