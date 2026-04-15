import { Alert, Button, Card, Col, Row, Space, Statistic, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import client from '../api/client'

type Overview = {
  total_requests: number
  active_users: number
  active_api_keys: number
  upstreams: number
}

export default function DashboardPage() {
  const navigate = useNavigate()
  const [data, setData] = useState<Overview>({
    total_requests: 0,
    active_users: 0,
    active_api_keys: 0,
    upstreams: 0
  })

  useEffect(() => {
    client.get('/admin/stats/overview').then((resp) => setData(resp.data)).catch(() => undefined)
  }, [])

  const readyForFirstCall = data.upstreams > 0 && data.active_api_keys > 0

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

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="快速开始">
            <Space direction="vertical" size={16} style={{ width: '100%' }}>
              <Alert
                type={readyForFirstCall ? 'success' : 'info'}
                showIcon
                message={readyForFirstCall ? '已经具备首次调用条件' : '还差一步就能发起第一次调用'}
                description={
                  readyForFirstCall
                    ? '你已经有上游和可用 Key，下一步可以直接复制示例请求进行联调。'
                    : '先完成“新建上游”和“生成 Key”，再用下方示例请求测试。'
                }
              />

              <div>
                <Typography.Title level={5}>1. 新建上游</Typography.Title>
                <Typography.Paragraph type="secondary">
                  选择一个服务商模板，填入官方 API Key。常见场景建议从 Google、Product Hunt、豆包 Embedding 开始。
                </Typography.Paragraph>
                <Button onClick={() => navigate('/upstreams')}>去配置上游</Button>
              </div>

              <div>
                <Typography.Title level={5}>2. 生成员工 Key</Typography.Title>
                <Typography.Paragraph type="secondary">
                  给内部调用方生成一个员工 Key。留空“允许访问的上游”表示该 Key 可以访问全部已启用上游。
                </Typography.Paragraph>
                <Button onClick={() => navigate('/api-keys')}>去生成 Key</Button>
              </div>

              <div>
                <Typography.Title level={5}>3. 发起第一次调用</Typography.Title>
                <Typography.Paragraph type="secondary">
                  调用方只需要记住网关地址、员工 Key，以及上游名称对应的路径。
                </Typography.Paragraph>
              </div>
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="示例请求">
            <Typography.Paragraph type="secondary">
              下面是一个通用调用格式。把 <code>{'{上游名}'}</code> 和请求体替换成实际值即可。
            </Typography.Paragraph>
            <pre
              style={{
                margin: 0,
                padding: 12,
                borderRadius: 8,
                background: '#0f172a',
                color: '#e2e8f0',
                overflowX: 'auto',
                fontSize: 13,
                lineHeight: 1.6
              }}
            >
{`curl http://localhost:8080/proxy/{上游名}/... \\
  -H "Authorization: Bearer sk-你的员工key" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"...","input":"..."}'`}
            </pre>
            <Space style={{ marginTop: 16 }}>
              <Button onClick={() => navigate('/logs')}>查看请求日志</Button>
              <Button onClick={() => navigate('/upstreams')}>查看上游路径说明</Button>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
