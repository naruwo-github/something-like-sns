import { auth0 } from "../../lib/auth0";

export default async function ProfilePage() {
  const session = await auth0.getSession();
  const user = session?.user as
    | { name?: string; email?: string; picture?: string }
    | undefined;

  if (!user) {
    return <p>Not authenticated</p>;
  }

  return (
    <div style={{ textAlign: "center" }}>
      {user.picture ? (
        // biome-ignore lint/performance/noImgElement: tmp
        <img
          alt={user.name || "profile"}
          src={user.picture}
          style={{ borderRadius: "50%", width: 80, height: 80 }}
        />
      ) : null}
      <h2>{user.name}</h2>
      <p>{user.email}</p>
      <pre>{JSON.stringify(user, null, 2)}</pre>
    </div>
  );
}
