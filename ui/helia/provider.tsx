"use client"
/* eslint-disable no-console */

import { UnixFS, unixfs } from '@helia/unixfs'
import { createHelia, Helia } from 'helia'
import PropTypes from 'prop-types'
import React, {
  useEffect,
  useState,
  useCallback,
  createContext
} from 'react'

export const HeliaContext = createContext({
  helia: null as (Helia | null),
  fs: null as (UnixFS | null),
  error: false as (boolean | null),
  starting: true
})

export const HeliaProvider = ({ children }: { children: React.ReactNode}) => {
  const [helia, setHelia] = useState<Helia | null>(null)
  const [fs, setFs] = useState<UnixFS | null>(null)
  const [starting, setStarting] = useState(true)
  const [error, setError] = useState<boolean | null>(null)

  const startHelia = useCallback(async () => {
    if (helia) {
      console.info('helia already started')
    } else if ((window as any).helia) {
      console.info('found a windowed instance of helia, populating ...')
      setHelia((window as any).helia)
      setFs(unixfs(helia!))
      setStarting(false)
    } else {
      try {
        console.info('Starting Helia')
        const helia = await createHelia()
        console.log('Helia started')
        console.log(helia)
        setHelia(helia)
        setFs(unixfs(helia))
        setStarting(false)
      } catch (e) {
        console.error(e)
        setError(true)
      }
    }
  }, [])

  useEffect(() => {
    startHelia()
  }, [])

  return (
    <HeliaContext.Provider
      value={{
        helia,
        fs,
        error,
        starting
      }}
    >{children}</HeliaContext.Provider>
  )
}

HeliaProvider.propTypes = {
  children: PropTypes.any
}