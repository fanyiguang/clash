import type{ AxiosError } from 'axios'
import classnames from 'classnames'
import { ResultAsync } from 'neverthrow'
import { useMemo, useLayoutEffect, useCallback } from 'react'

import EE, { Action } from '@lib/event'
import { isClashX, jsBridge } from '@lib/jsBridge'
import { Proxy as IProxy } from '@lib/request'
import { BaseComponentProps } from '@models'
import { useClient, useProxy } from '@stores'

import './style.scss'

import toast from 'react-hot-toast';

interface ProxyProps extends BaseComponentProps {
    config: IProxy
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

export function Proxy (props: ProxyProps) {
    const { config, className } = props
    const { set } = useProxy()
    const client = useClient()

    const getDelay = useCallback(async (name: string) => {
        if (isClashX()) {
            const delay = await jsBridge?.getProxyDelay(name) ?? 0
            return delay
        }

        const { data: { delay } } = await client.getProxyDelay(name)
        return delay
    }, [client])

    const speedTest = useCallback(async function () {
        const result = await ResultAsync.fromPromise(getDelay(config.name), e => e as AxiosError)

        const validDelay = result.isErr() ? 0 : result.value
        set(draft => {
            const proxy = draft.proxies.find(p => p.name === config.name)
            if (proxy != null) {
                proxy.history.push({ time: Date.now().toString(), delay: validDelay })
            }
        })
    }, [config.name, getDelay, set])

    const delay = useMemo(
        () => config.history?.length ? config.history.slice(-1)[0].delay : 0,
        [config],
    )

    useLayoutEffect(() => {
        const handler = () => { speedTest() }
        EE.subscribe(Action.SPEED_NOTIFY, handler)
        return () => EE.unsubscribe(Action.SPEED_NOTIFY, handler)
    }, [speedTest])

    const hasError = useMemo(() => delay === 0, [delay])
    const color = useMemo(
        () => Object.keys(TagColors).find(
            threshold => delay <= TagColors[threshold as keyof typeof TagColors],
        ),
        [delay],
    )

    const backgroundColor = hasError ? '#E5E7EB' : color

    function copyProxyInfo() {
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



    const unreadMessages = props.config?.all ?? []

    return (
        <div className={classnames('proxy-item', { 'opacity-50': hasError }, className)}>

            <div className="flex-1 ">
                <div className="flex items-center display:flex">
                    <span
                        className={classnames('rounded-sm py-[3px] px-1 text-[10px] text-white', { 'text-gray-600': hasError })}
                        style={{ backgroundColor }}>
                        {config.type}
                    </span>
                    <div className="proxy-info-copy flex flex-auto items-center justify-end">
                            <p className="rounded-sm py-[3px] px-1 text-[10px]"
                               style={{
                                   backgroundColor: '#E5E7EB',
                               }}
                            onClick={copyProxyInfo}>info</p>
                    </div>
                </div>
                <p className="proxy-name">{config.name}</p>

                {unreadMessages.map((item, index) => {
                    return(
                    <div className="relay-node">
                       - {item}
                    </div>)
                })}

            </div>

            <div className="flex h-full flex-col items-center justify-center space-y-3 text-[10px] md:h-[18px] md:flex-row md:justify-between md:space-y-0">
                <p>{delay === 0 ? '-' : `${delay}ms`}</p>
                { config.udp && <p className="rounded bg-gray-200 p-[3px] text-gray-600">UDP</p> }
            </div>
        </div>
    )
}

