import { Card, Col, Row, Statistic } from 'antd'
import { useEffect, useState } from 'react'
import client from '../api/client'

type Overview = {
  total_requests: number
  active_users: number
  active_api_keys: number
  upstreams: number
}

export default function DashboardPage() {
  const [data, setData] = useState<Overview>({
    total_requests: 0,
    active_users: 0,
    active_api_keys: 0,
    upstreams: 0
  })

  useEffect(() => {
    client.get('/admin/stats/overview').then((resp) => setData(resp.data)).catch(() => undefined)
  }, [])

  return (
    <div>
      <Row gutter={16}>
        <Col span={6}>
          <Card>
            <Statistic title="总请求数" value={data.total_requests} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="活跃用户" value={data.active_users} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="活跃 Key" value={data.active_api_keys} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="上游数量" value={data.upstreams} />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
