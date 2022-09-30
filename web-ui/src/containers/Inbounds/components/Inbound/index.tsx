import type{ AxiosError } from 'axios'
import classnames from 'classnames'
import { ResultAsync } from 'neverthrow'
import { useMemo, useLayoutEffect, useCallback } from 'react'

import EE, { Action } from '@lib/event'
import { isClashX, jsBridge } from '@lib/jsBridge'
import {Inbound as AInbound} from '@lib/request'
import { BaseComponentProps } from '@models'
import { useClient, useProxy } from '@stores'

import './style.scss'

import toast from 'react-hot-toast';

interface ProxyProps extends BaseComponentProps {
    config: AInbound
}

interface RelayNodeProps extends BaseComponentProps {
    nodes: string[]
}

const TagColors = {
    '#909399': 0,
    '#00c520': 260,
    '#ff9a28': 600,
    '#ff3e5e': Infinity,
}

export function Inbound (props: ProxyProps) {

    const { config, className } = props

    const backgroundColor = '#00C520'

    function copyInboundInfo() {
        if (props.config.secretData == undefined) {
            toast.error("No data",{duration: 500});
            return;
        }

        navigator.clipboard.writeText(props.config.secretData).catch(
            // navigator.clipboard.writeText("ssss").catch(
            () => toast('err', { duration: 500 })
        ).then(
            () => toast('ok',{ duration: 500 })
        )
    }

    return (
        <div className={classnames('inbound-item', config.type)}>

            <div className="flex-1 ">
                <div className="flex items-center display:flex">
                    <span
                        className={classnames('rounded-sm py-[3px] px-1 text-[10px] text-white')}
                        style={{ backgroundColor }}>
                        {config.type}
                    </span>
                    <div className="proxy-info-copy flex flex-auto items-center justify-end">
                        <p className="rounded-sm py-[3px] px-1 text-[10px]"
                           style={{
                               backgroundColor: '#E5E7EB',
                           }}
                           onClick={copyInboundInfo}>info</p>
                    </div>
                </div>
                <p className="inbound-name">{config.name}</p>

                <p className="inbound-addr">{config.rawAddress}</p>

            </div>


        </div>
    )
}

