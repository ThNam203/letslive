"use client";

import { IconEmail } from "@/components/icons/email";
import { IconEye } from "@/components/icons/eye";
import { IconEyeOff } from "@/components/icons/eye-off";
import { IconPasswordOutline } from "@/components/icons/password";
import GLOBAL from "@/global";
import { Spinner } from "@nextui-org/spinner";
import { useState } from "react";
import { toast } from "react-toastify";

export default function LogInForm() {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [hidingPassword, setHidingPassword] = useState(true);
    const [isLoading, setIsLoading] = useState(false);
    const [errors, setErrors] = useState({
        email: "",
        password: "",
    });

    const validate = () => {
        const newErrors = { email: "", password: "" };

        if (!email) {
            newErrors.email = "Email is required";
        } else if (!/\S+@\S+\.\S+/.test(email)) {
            newErrors.email = "Email is invalid";
        }

        if (!password) {
            newErrors.password = "Password is required";
        } else if (password.length < 8) {
            newErrors.password = "Password must be at least 8 characters";
        }

        setErrors(newErrors);

        return !newErrors.email && !newErrors.password;
    };

    const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        e.stopPropagation();

        if (validate()) {
            setIsLoading(true);
            fetch(GLOBAL.API_URL + "/v1/auth/login", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ email, password }),
            })
                .catch((err) => {
                    toast.error(err.message);
                })
                .finally(() => setIsLoading(false));
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <div className="flex px-4 gap-4 items-center rounded-md border border-gray-200">
                <label htmlFor="email">
                    <IconEmail className="opacity-40 scale-125" />
                </label>
                <input
                    id="email"
                    aria-label="Email"
                    className="h-[50px] focus:outline-none flex-1"
                    placeholder="Email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                />
            </div>
            {errors.email && (
                <p className="text-red-500 text-xs font-semibold">
                    {errors.email}
                </p>
            )}
            <div className="flex px-4 gap-4 items-center rounded-md border border-gray-200 mt-4">
                <label htmlFor="password">
                    <IconPasswordOutline className="opacity-40 scale-125" />
                </label>
                <input
                    id="password"
                    aria-label="Password"
                    className="h-[50px] focus:outline-none flex-1"
                    placeholder="Password"
                    type={hidingPassword ? "password" : "text"}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                />
                {!hidingPassword ? (
                    <IconEye
                        className="scale-150 opacity-50"
                        onClick={() => setHidingPassword(true)}
                    />
                ) : (
                    <IconEyeOff
                        className="scale-150 opacity-50"
                        onClick={() => setHidingPassword(false)}
                    />
                )}
            </div>
            {errors.password && (
                <p className="text-red-500 text-xs font-semibold">
                    {errors.password}
                </p>
            )}
            <button
                type="submit"
                disabled={isLoading}
                className="w-full rounded-md flex justify-center items-center bg-blue-400 hover:bg-blue-500 text-white h-[50px] border-transparent border mt-4 font-semibold"
            >
                LOG IN
                {isLoading && <Spinner className="ml-2" />}
            </button>
        </form>
    );
}
