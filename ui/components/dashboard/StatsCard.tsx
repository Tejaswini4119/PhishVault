interface StatsCardProps {
    title: string;
    value: string | number;
    change?: string;
    trend?: 'up' | 'down' | 'neutral';
    icon?: React.ReactNode;
}

const StatsCard = ({ title, value, change, trend, icon }: StatsCardProps) => {
    const trendColor = trend === 'up' ? 'text-green-400' : trend === 'down' ? 'text-red-400' : 'text-slate-400';

    return (
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-6 hover:border-slate-700 transition-colors">
            <div className="flex justify-between items-start mb-4">
                <div>
                    <h3 className="text-slate-400 text-sm font-medium">{title}</h3>
                    <p className="text-2xl font-bold text-white mt-1">{value}</p>
                </div>
                {icon && <div className="p-2 bg-slate-800 rounded-lg text-blue-400">{icon}</div>}
            </div>

            {change && (
                <div className="flex items-center text-xs">
                    <span className={`${trendColor} font-medium mr-2`}>
                        {trend === 'up' ? '↑' : trend === 'down' ? '↓' : '•'} {change}
                    </span>
                    <span className="text-slate-500">vs last week</span>
                </div>
            )}
        </div>
    );
};

export default StatsCard;
