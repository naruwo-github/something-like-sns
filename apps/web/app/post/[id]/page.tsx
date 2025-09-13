import PostDetailContainer from "../_containers/post-detail/container";

export default async function Page({ params }: { params: { id: string } }) {
  const postId = Number(params.id);
  return <PostDetailContainer postId={postId} />;
}
