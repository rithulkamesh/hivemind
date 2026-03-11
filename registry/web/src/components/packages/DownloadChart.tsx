import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

interface DataPoint {
  date: string;
  downloads: number;
}

interface DownloadChartProps {
  data: DataPoint[];
}

export function DownloadChart({ data }: DownloadChartProps) {
  return (
    <div className="h-64 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#1f1f1f" />
          <XAxis
            dataKey="date"
            tick={{ fill: "#666", fontSize: 11 }}
            stroke="#1f1f1f"
          />
          <YAxis
            tick={{ fill: "#666", fontSize: 11 }}
            stroke="#1f1f1f"
          />
          <Tooltip
            contentStyle={{
              background: "#111",
              border: "1px solid #1f1f1f",
              borderRadius: 0,
            }}
            labelStyle={{ color: "#fafafa" }}
          />
          <Line
            type="monotone"
            dataKey="downloads"
            stroke="#F5A623"
            strokeWidth={2}
            dot={{ fill: "#F5A623", r: 2 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
