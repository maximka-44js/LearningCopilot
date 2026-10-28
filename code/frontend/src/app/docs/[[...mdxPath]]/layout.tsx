import { Footer, Layout, Navbar } from "nextra-theme-docs";
import { getPageMap } from "nextra/page-map";
import "nextra-theme-docs/style.css";

export default async function DocsLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const pageMap = await getPageMap("/docs");

  return (
    <Layout
      navbar={<Navbar logo={<b>LC Frontend Docs</b>} />}
      pageMap={pageMap}
      docsRepositoryBase="https://github.com"
      footer={<Footer>MIT {new Date().getFullYear()} © LC Frontend.</Footer>}
      editLink={"Edit this page"}
      feedback={{ content: "Question? Give us feedback ->" }}
    >
      {children}
    </Layout>
  );
}
