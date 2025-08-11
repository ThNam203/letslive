"use client";

import { useState, useEffect } from 'react'
import { useTheme } from 'next-themes'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../ui/select';
import { cn } from '@/utils/cn';
import { THEME_COLORS } from '@/constant/theme';


const ThemeSwitch = ({
  className,
}: {
  className?: string
}) => {
  const [mounted, setMounted] = useState(false)
  const { theme, setTheme } = useTheme()

  // useEffect only runs on the client, so now we can safely show the UI
  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return null
  }

  return (
      <Select onValueChange={(value) => setTheme(value)} value={theme}>
        <SelectTrigger className={cn("w-fit border-border", className)} >
          <SelectValue defaultValue={THEME_COLORS.SYSTEM} />
        </SelectTrigger>
        <SelectContent className='bg-background text-foreground border-border border'>
          <SelectGroup>
            {Object.values(THEME_COLORS).map((color) => (
              <SelectItem key={color} value={color}>
                {color.charAt(0).toUpperCase() + color.slice(1)}
              </SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>
  )
}

export default ThemeSwitch