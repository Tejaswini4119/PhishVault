"use client";

import { useState } from "react";
import { useAuth } from "@/components/AuthProvider";
import Link from "next/link";

export default function LoginPage() {
    const { login } = useAuth();
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [isRegister, setIsRegister] = useState(false);
    const [error, setError] = useState("");

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");

        const endpoint = isRegister ? "/api/auth/register" : "/api/auth/login";

        try {
            const res = await fetch(endpoint, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ username, password }),
            });

            if (!res.ok) {
                const data = await res.text(); // or json depending on error response
                throw new Error(data || "Request failed");
            }

            if (isRegister) {
                // If register success, switch to login or auto-login
                // For simplified flow, just switch to login
                setIsRegister(false);
                setError("Registration successful! Please login.");
            } else {
                const data = await res.json();
                login(data.token, data.username);
            }
        } catch (err: any) {
            // Check if it is a JSON error message first
            try {
                const jsonErr = JSON.parse(err.message);
                setError(jsonErr.message || "Failed");
            } catch {
                setError(err.message);
            }
        }
    };

    return (
        <div className="flex flex-col items-center justify-center min-h-[80vh]">
            <div className="bg-slate-900 p-8 rounded-xl border border-slate-800 shadow-2xl w-full max-w-md">
                <h1 className="text-3xl font-bold text-white mb-6 text-center">
                    {isRegister ? "Create Account" : "Analyst Login"}
                </h1>

                {error && (
                    <div className="bg-red-900/50 border border-red-800 text-red-200 px-4 py-2 rounded mb-4 text-sm">
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-slate-400 text-sm font-medium mb-1">Username</label>
                        <input
                            type="text"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            className="w-full bg-slate-950 border border-slate-800 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500 transition-colors"
                            placeholder="username"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-slate-400 text-sm font-medium mb-1">Password</label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            className="w-full bg-slate-950 border border-slate-800 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500 transition-colors"
                            placeholder="••••••••"
                            required
                        />
                    </div>

                    <button
                        type="submit"
                        className="w-full bg-blue-600 hover:bg-blue-500 text-white font-bold py-2 px-4 rounded-lg transition-colors mt-2"
                    >
                        {isRegister ? "Register" : "Sign In"}
                    </button>
                </form>

                <div className="mt-6 text-center text-sm text-slate-500">
                    {isRegister ? "Already have an account?" : "Need an account?"}{" "}
                    <button
                        onClick={() => {
                            setIsRegister(!isRegister);
                            setError("");
                        }}
                        className="text-blue-400 hover:text-blue-300 font-medium"
                    >
                        {isRegister ? "Sign In" : "Register"}
                    </button>
                </div>
            </div>
        </div>
    );
}
