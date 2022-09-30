import { useMemo } from 'react'

import { Card, Header, Icon, Checkbox } from '@components'
import EE from '@lib/event'
import { useRound } from '@lib/hook'
import * as API from '@lib/request'
import {useI18n, useConfig, useProxy, useProxyProviders, useGeneral, useInbounds} from '@stores'

import { Inbound } from './components'
import './style.scss'
import {Proxy} from "@containers/Proxies/components";
import useSWR from "swr";


export default function Inbounds () {

    const { inbounds,update}  = useInbounds()

    useSWR('inbounds', update)

    console.log(inbounds)

    return (
        <div className="page">
         {
             inbounds.length !== 0 &&
                 <div className="flex flex-col">
                     <ul className="inbounds-list">
                         {
                             inbounds.map(p => (
                                 <li>
                                 <Inbound  config={p} />
                                 </li>
                             ))
                         }
                     </ul>
                 </div>
         }
        </div>
    )
}
