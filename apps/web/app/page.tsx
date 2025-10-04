export const dynamic = "force-dynamic";
export const fetchCache = "default-no-store";

import FeedContainer from "app/_containers/feed/container";

export default async function Page() {
  return <FeedContainer />;
}
