export default function Loading() {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
      <div
        style={{
          height: 28,
          width: 160,
          background: "#f3f3f3",
          borderRadius: 6,
        }}
      />
      <div style={{ display: "flex", gap: 8 }}>
        <div
          style={{
            flex: 1,
            height: 36,
            background: "#f3f3f3",
            borderRadius: 6,
          }}
        />
        <div
          style={{
            width: 72,
            height: 36,
            background: "#f3f3f3",
            borderRadius: 6,
          }}
        />
      </div>
      <ul style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {["s1", "s2", "s3"].map((k) => (
          <li
            key={k}
            style={{ border: "1px solid #eee", padding: 12, borderRadius: 8 }}
          >
            <div
              style={{
                height: 14,
                width: 220,
                background: "#f3f3f3",
                borderRadius: 4,
                marginBottom: 8,
              }}
            />
            <div
              style={{
                height: 16,
                width: "100%",
                background: "#f3f3f3",
                borderRadius: 4,
                marginBottom: 6,
              }}
            />
            <div
              style={{
                height: 16,
                width: "80%",
                background: "#f3f3f3",
                borderRadius: 4,
              }}
            />
          </li>
        ))}
      </ul>
    </div>
  );
}
