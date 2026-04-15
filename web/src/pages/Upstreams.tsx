import {
  Alert,
  Button,
  Card,
  Collapse,
  Descriptions,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Typography,
  message
} from 'antd'
import { useEffect, useMemo, useState } from 'react'
import client from '../api/client'

type Upstream = {
  id: number
  name: string
  display_name: string
  base_url: string
  auth_type: string
  auth_key: string
  auth_value: string
  timeout_seconds: number
  proxy_url: string
  strip_prefix: boolean
  extra_headers: string
  is_active: boolean
  description: string
}

type UpstreamTestResult = {
  ok: boolean
  reachable: boolean
  category: string
  message: string
  target_url: string
  status_code?: number
  response_excerpt?: string
}

type UpstreamTemplateKey = 'doubao' | 'openai' | 'deepseek' | 'google' | 'producthunt'

type UpstreamTemplate = {
  key: UpstreamTemplateKey
  title: string
  summary: string
  authLabel: string
  requestMethod: 'GET' | 'POST'
  requestPath: string
  requestBody?: string
  formValues: Partial<Upstream>
}

const defaultFormValues: Partial<Upstream> = {
  auth_type: 'none',
  auth_key: '',
  auth_value: '',
  timeout_seconds: 120,
  proxy_url: '',
  strip_prefix: true,
  extra_headers: '{}',
  is_active: true,
  description: ''
}

const upstreamTemplates: UpstreamTemplate[] = [
  {
    key: 'doubao',
    title: '豆包 Embedding',
    summary: '适合火山方舟 Embedding / 多模态向量场景，Bearer 鉴权。',
    authLabel: 'Bearer',
    requestMethod: 'POST',
    requestPath: '/api/v3/embeddings/multimodal',
    requestBody: JSON.stringify(
      {
        model: 'doubao-embedding-vision-251215',
        input: [{ type: 'text', text: '这是一个测试文本' }],
        dimensions: 2048
      },
      null,
      2
    ),
    formValues: {
      name: 'doubao_embedding',
      display_name: 'Doubao Embedding',
      base_url: 'https://ark.cn-beijing.volces.com',
      auth_type: 'bearer',
      auth_key: '',
      auth_value: '',
      description: '豆包 Embedding，多模态向量接口。官方 API Key 填在 auth_value，dimensions 常用 2048。'
    }
  },
  {
    key: 'openai',
    title: 'OpenAI Compatible',
    summary: '标准 OpenAI 兼容接口，适合大模型聊天与 Embedding。',
    authLabel: 'Bearer',
    requestMethod: 'POST',
    requestPath: '/v1/chat/completions',
    requestBody: JSON.stringify(
      {
        model: 'gpt-4.1-mini',
        messages: [{ role: 'user', content: 'say hello' }]
      },
      null,
      2
    ),
    formValues: {
      name: 'openai',
      display_name: 'OpenAI Compatible',
      base_url: 'https://api.openai.com',
      auth_type: 'bearer',
      auth_key: '',
      auth_value: '',
      description: 'OpenAI 兼容接口。若是其他兼容厂商，可只改 base_url 和模型名。'
    }
  },
  {
    key: 'deepseek',
    title: 'DeepSeek',
    summary: 'DeepSeek 官方接口，Bearer 鉴权，路径与 OpenAI 兼容。',
    authLabel: 'Bearer',
    requestMethod: 'POST',
    requestPath: '/chat/completions',
    requestBody: JSON.stringify(
      {
        model: 'deepseek-chat',
        messages: [{ role: 'user', content: '写一个 hello world' }]
      },
      null,
      2
    ),
    formValues: {
      name: 'deepseek',
      display_name: 'DeepSeek',
      base_url: 'https://api.deepseek.com',
      auth_type: 'bearer',
      auth_key: '',
      auth_value: '',
      description: 'DeepSeek 官方接口。若使用兼容模式，也可沿用 OpenAI Compatible 模板。'
    }
  },
  {
    key: 'google',
    title: 'Google API',
    summary: '适合 query 参数带 key 的接口，例如部分 Google API。',
    authLabel: 'Query',
    requestMethod: 'GET',
    requestPath: '/youtube/v3/search?part=snippet&q=golang',
    formValues: {
      name: 'google',
      display_name: 'Google API',
      base_url: 'https://www.googleapis.com',
      auth_type: 'query',
      auth_key: 'key',
      auth_value: '',
      description: 'Google API。官方 API Key 填在 auth_value，系统会自动追加到 query 参数。'
    }
  },
  {
    key: 'producthunt',
    title: 'Product Hunt GraphQL',
    summary: 'Product Hunt GraphQL 场景，Bearer 鉴权。',
    authLabel: 'Bearer',
    requestMethod: 'POST',
    requestPath: '/v2/api/graphql',
    requestBody: JSON.stringify(
      {
        query: 'query { viewer { user { name } } }'
      },
      null,
      2
    ),
    formValues: {
      name: 'producthunt',
      display_name: 'Product Hunt GraphQL',
      base_url: 'https://api.producthunt.com',
      auth_type: 'bearer',
      auth_key: '',
      auth_value: '',
      description: 'Product Hunt GraphQL API。保存后可直接复制 GraphQL 示例请求。'
    }
  }
]

function getErrorMessage(err: any): string {
  return err?.response?.data?.error || err?.message || '请求失败'
}

function matchTemplate(upstream: Partial<Upstream>): UpstreamTemplate | null {
  const fingerprint = `${upstream.name || ''} ${upstream.display_name || ''} ${upstream.base_url || ''}`.toLowerCase()
  if (fingerprint.includes('ark.cn-beijing.volces.com') || fingerprint.includes('doubao') || fingerprint.includes('volces')) {
    return upstreamTemplates.find((item) => item.key === 'doubao') || null
  }
  if (fingerprint.includes('api.deepseek.com') || fingerprint.includes('deepseek')) {
    return upstreamTemplates.find((item) => item.key === 'deepseek') || null
  }
  if (fingerprint.includes('api.openai.com') || fingerprint.includes('openai')) {
    return upstreamTemplates.find((item) => item.key === 'openai') || null
  }
  if (fingerprint.includes('googleapis.com') || fingerprint.includes('google')) {
    return upstreamTemplates.find((item) => item.key === 'google') || null
  }
  if (fingerprint.includes('producthunt.com') || fingerprint.includes('producthunt')) {
    return upstreamTemplates.find((item) => item.key === 'producthunt') || null
  }
  return null
}

function buildAuthSummary(upstream: Partial<Upstream>): string {
  if (upstream.auth_type === 'none') {
    return '上游不带额外鉴权信息。注意：当前网关不会透传调用方 Authorization 到上游。'
  }
  if (upstream.auth_type === 'bearer') {
    return '网关会把 auth_value 作为 Bearer Token 发给上游。'
  }
  if (upstream.auth_type === 'header') {
    return `网关会把 auth_value 写入请求头 ${upstream.auth_key || '(请填写 auth_key)'}。`
  }
  if (upstream.auth_type === 'query') {
    return `网关会把 auth_value 追加为 query 参数 ${upstream.auth_key || '(请填写 auth_key)'}。`
  }
  return '未识别的鉴权方式。'
}

function buildCurlExample(upstream: Upstream): string {
  const template = matchTemplate(upstream)
  const path = template?.requestPath || '/v1/chat/completions'
  const method = template?.requestMethod || 'POST'
  const proxyPath = `/proxy/${upstream.name}${path.startsWith('/') ? path : `/${path}`}`

  const lines = [`curl http://localhost:8080${proxyPath} \\`]
  if (method !== 'GET') {
    lines.push(`  -X ${method} \\`)
  }
  lines.push('  -H "Authorization: Bearer sk-你的员工key" \\')

  if (method === 'POST') {
    lines.push('  -H "Content-Type: application/json" \\')
    lines.push(`  -d '${template?.requestBody || JSON.stringify({ model: '...', input: '...' }, null, 2)}'`)
  }

  return lines.join('\n')
}

function getAuthValueLabel(authType?: string): string {
  if (authType === 'bearer') return '上游 Bearer Token'
  if (authType === 'header') return 'header 值'
  if (authType === 'query') return 'query 参数值'
  return 'auth_value'
}

function getAuthValuePlaceholder(authType?: string): string {
  if (authType === 'bearer') return '填写官方 API Key 或 Token'
  if (authType === 'header') return '例如 sk-xxx 或其他 header 值'
  if (authType === 'query') return '例如 Google API Key'
  return '鉴权值'
}

export default function UpstreamsPage() {
  const [items, setItems] = useState<Upstream[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Upstream | null>(null)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [testingID, setTestingID] = useState<number | null>(null)
  const [testResult, setTestResult] = useState<UpstreamTestResult | null>(null)
  const [exampleTarget, setExampleTarget] = useState<Upstream | null>(null)
  const [form] = Form.useForm()
  const authType = Form.useWatch('auth_type', form)
  const authKey = Form.useWatch('auth_key', form)

  const exampleText = useMemo(() => {
    if (!exampleTarget) {
      return ''
    }
    return buildCurlExample(exampleTarget)
  }, [exampleTarget])

  const load = async () => {
    setLoading(true)
    try {
      const resp = await client.get('/admin/upstreams')
      setItems(resp.data.items || [])
    } catch (err: any) {
      message.error(`加载上游失败：${getErrorMessage(err)}`)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load().catch(() => undefined)
  }, [])

  const closeEditor = () => {
    setOpen(false)
    setEditing(null)
    form.resetFields()
  }

  const applyTemplate = (templateKey: UpstreamTemplateKey) => {
    const template = upstreamTemplates.find((item) => item.key === templateKey)
    if (!template) {
      return
    }
    form.setFieldsValue({
      ...defaultFormValues,
      ...template.formValues
    })
  }

  const openCreate = (templateKey?: UpstreamTemplateKey) => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue(defaultFormValues)
    if (templateKey) {
      applyTemplate(templateKey)
    }
    setOpen(true)
  }

  const openEdit = (item: Upstream) => {
    setEditing(item)
    form.setFieldsValue({
      ...defaultFormValues,
      ...item
    })
    setOpen(true)
  }

  const submit = async () => {
    setSubmitting(true)
    try {
      const values = await form.validateFields()
      if (editing) {
        await client.put(`/admin/upstreams/${editing.id}`, values)
        message.success('更新成功')
      } else {
        await client.post('/admin/upstreams', values)
        message.success('创建成功')
      }
      closeEditor()
      await load()
    } catch (err: any) {
      message.error(`保存失败：${getErrorMessage(err)}`)
    } finally {
      setSubmitting(false)
    }
  }

  const del = async (id: number) => {
    setSubmitting(true)
    try {
      await client.delete(`/admin/upstreams/${id}`)
      message.success('删除成功')
      await load()
    } catch (err: any) {
      message.error(`删除失败：${getErrorMessage(err)}`)
    } finally {
      setSubmitting(false)
    }
  }

  const testUpstream = async (item: Upstream) => {
    setTestingID(item.id)
    try {
      const resp = await client.post('/admin/upstreams/test', item)
      setTestResult(resp.data)
    } catch (err: any) {
      message.error(`测试失败：${getErrorMessage(err)}`)
    } finally {
      setTestingID(null)
    }
  }

  const emptyState = (
    <Empty
      image={Empty.PRESENTED_IMAGE_SIMPLE}
      description="还没有上游。建议先套用模板，再填官方 API Key。"
    >
      <Space wrap>
        <Button type="primary" onClick={() => openCreate('doubao')}>
          新建豆包上游
        </Button>
        <Button onClick={() => openCreate('openai')}>新建 OpenAI 兼容上游</Button>
        <Button onClick={() => openCreate()}>自定义上游</Button>
      </Space>
    </Empty>
  )

  return (
    <Card
      title="上游 API 管理"
      extra={
        <Button type="primary" onClick={() => openCreate()}>
          新建上游
        </Button>
      }
    >
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Alert
          type="info"
          showIcon
          message="推荐配置顺序"
          description="1. 先选服务商模板  2. 填官方 API Key  3. 先点测试确认连通  4. 保存后复制专属调用示例给调用方。"
        />

        <Table
          rowKey="id"
          loading={loading}
          dataSource={items}
          locale={{ emptyText: emptyState }}
          columns={[
            {
              title: '上游',
              render: (_, row: Upstream) => (
                <Space direction="vertical" size={2}>
                  <Typography.Text strong>{row.display_name || row.name}</Typography.Text>
                  <Typography.Text type="secondary">{row.name}</Typography.Text>
                </Space>
              )
            },
            {
              title: 'Base URL',
              dataIndex: 'base_url',
              render: (value: string) => (
                <Typography.Text copyable={{ text: value }} style={{ maxWidth: 280 }}>
                  {value}
                </Typography.Text>
              )
            },
            {
              title: '鉴权',
              render: (_, row: Upstream) => <Tag>{row.auth_type}</Tag>
            },
            {
              title: '状态',
              dataIndex: 'is_active',
              render: (value: boolean) => <Tag color={value ? 'green' : 'red'}>{value ? '启用' : '停用'}</Tag>
            },
            {
              title: '说明',
              dataIndex: 'description',
              render: (value: string) => (
                <Typography.Text type="secondary">
                  {value || '未填写说明'}
                </Typography.Text>
              )
            },
            {
              title: '操作',
              render: (_, row: Upstream) => (
                <Space wrap>
                  <Button size="small" loading={testingID === row.id} onClick={() => testUpstream(row)}>
                    测试
                  </Button>
                  <Button size="small" onClick={() => setExampleTarget(row)}>
                    示例
                  </Button>
                  <Button size="small" onClick={() => openEdit(row)}>
                    编辑
                  </Button>
                  <Button danger size="small" loading={submitting} onClick={() => del(row.id)}>
                    删除
                  </Button>
                </Space>
              )
            }
          ]}
        />
      </Space>

      <Modal
        open={open}
        width={760}
        onCancel={closeEditor}
        onOk={submit}
        confirmLoading={submitting}
        title={editing ? '编辑上游' : '新建上游'}
      >
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          {!editing && (
            <>
              <Typography.Title level={5} style={{ margin: 0 }}>
                先选一个最接近的模板
              </Typography.Title>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 12 }}>
                {upstreamTemplates.map((template) => (
                  <div
                    key={template.key}
                    onClick={() => applyTemplate(template.key)}
                    style={{
                      border: '1px solid #d9d9d9',
                      borderRadius: 8,
                      padding: 12,
                      cursor: 'pointer',
                      background: '#fafafa'
                    }}
                  >
                    <Space direction="vertical" size={6} style={{ width: '100%' }}>
                      <Space wrap>
                        <Typography.Text strong>{template.title}</Typography.Text>
                        <Tag color="blue">{template.authLabel}</Tag>
                      </Space>
                      <Typography.Text type="secondary">{template.summary}</Typography.Text>
                      <Typography.Text type="secondary">示例路径：{template.requestPath}</Typography.Text>
                    </Space>
                  </div>
                ))}
              </div>
            </>
          )}

          <Alert
            type={authType === 'none' ? 'warning' : 'info'}
            showIcon
            message={authType === 'none' ? '当前选择为无鉴权' : '当前上游鉴权说明'}
            description={buildAuthSummary({ auth_type: authType, auth_key: authKey })}
          />

          <Form form={form} layout="vertical">
            <Form.Item label="name" name="name" rules={[{ required: true }]}>
              <Input placeholder="openai / deepseek / doubao_embedding / google" />
            </Form.Item>
            <Form.Item label="display_name" name="display_name" rules={[{ required: true }]}>
              <Input placeholder="给后台管理员看的显示名称" />
            </Form.Item>
            <Form.Item label="base_url" name="base_url" rules={[{ required: true }]}>
              <Input placeholder="例如 https://api.openai.com" />
            </Form.Item>
            <Form.Item
              label="auth_type"
              name="auth_type"
              rules={[{ required: true }]}
              extra="调用方访问网关时始终使用员工 Key；这里配置的是网关转发到上游时如何带官方鉴权信息。"
            >
              <Select
                options={[
                  { value: 'none', label: 'none' },
                  { value: 'bearer', label: 'bearer' },
                  { value: 'header', label: 'header' },
                  { value: 'query', label: 'query' }
                ]}
              />
            </Form.Item>
            {(authType === 'bearer' || authType === 'header' || authType === 'query') && (
              <Form.Item label={getAuthValueLabel(authType)} name="auth_value">
                <Input.Password placeholder={getAuthValuePlaceholder(authType)} />
              </Form.Item>
            )}
            <Form.Item label="description" name="description">
              <Input.TextArea rows={2} placeholder="记录用途、模型名、建议路径等，方便后续维护" />
            </Form.Item>
            <Form.Item label="is_active" name="is_active" valuePropName="checked">
              <Switch />
            </Form.Item>

            <Collapse
              ghost
              items={[
                {
                  key: 'advanced',
                  label: '高级设置',
                  children: (
                    <Space direction="vertical" size={0} style={{ width: '100%' }}>
                      {(authType === 'header' || authType === 'query') && (
                        <Form.Item
                          label="auth_key"
                          name="auth_key"
                          rules={[
                            {
                              required: true,
                              message: `auth_type 为 ${authType} 时 auth_key 不能为空`
                            }
                          ]}
                        >
                          <Input placeholder="query/header 模式必填，例如 key / x-api-key / Authorization" />
                        </Form.Item>
                      )}
                      <Form.Item label="timeout_seconds" name="timeout_seconds">
                        <InputNumber min={1} style={{ width: '100%' }} />
                      </Form.Item>
                      <Form.Item label="proxy_url" name="proxy_url">
                        <Input placeholder="需要代理时填写，例如 http://127.0.0.1:7890" />
                      </Form.Item>
                      <Form.Item label="strip_prefix" name="strip_prefix" valuePropName="checked">
                        <Switch />
                      </Form.Item>
                      <Form.Item label="extra_headers(JSON)" name="extra_headers">
                        <Input.TextArea rows={3} placeholder='例如 {"x-foo":"bar"}' />
                      </Form.Item>
                    </Space>
                  )
                }
              ]}
            />
          </Form>
        </Space>
      </Modal>

      <Modal open={!!testResult} footer={null} onCancel={() => setTestResult(null)} title="测试连接结果">
        {testResult && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Alert
              type={testResult.ok ? 'success' : testResult.reachable ? 'warning' : 'error'}
              showIcon
              message={testResult.message}
              description={
                testResult.status_code
                  ? `状态码：${testResult.status_code}，分类：${testResult.category}`
                  : `分类：${testResult.category}`
              }
            />
            <div>
              <Typography.Text type="secondary">目标地址</Typography.Text>
              <Typography.Paragraph copyable style={{ marginBottom: 0 }}>
                {testResult.target_url}
              </Typography.Paragraph>
            </div>
            <div>
              <Typography.Text type="secondary">响应摘要</Typography.Text>
              <pre
                style={{
                  margin: '8px 0 0',
                  padding: 12,
                  borderRadius: 8,
                  background: '#f5f5f5',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word'
                }}
              >
                {testResult.response_excerpt || '无响应体'}
              </pre>
            </div>
          </Space>
        )}
      </Modal>

      <Modal open={!!exampleTarget} footer={null} onCancel={() => setExampleTarget(null)} title="调用示例">
        {exampleTarget && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Descriptions
              bordered
              size="small"
              column={1}
              items={[
                {
                  key: 'name',
                  label: '上游名称',
                  children: `${exampleTarget.display_name || exampleTarget.name} (${exampleTarget.name})`
                },
                {
                  key: 'route',
                  label: '网关路径前缀',
                  children: `/proxy/${exampleTarget.name}/...`
                },
                {
                  key: 'auth',
                  label: '上游鉴权方式',
                  children: buildAuthSummary(exampleTarget)
                }
              ]}
            />

            <div>
              <Typography.Text type="secondary">专属 curl 示例</Typography.Text>
              <Typography.Paragraph
                copyable={{ text: exampleText, tooltips: ['复制', '已复制'] }}
                style={{
                  marginTop: 8,
                  marginBottom: 0,
                  padding: 12,
                  borderRadius: 8,
                  background: '#0f172a',
                  color: '#e2e8f0',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  fontFamily: 'monospace'
                }}
              >
                {exampleText}
              </Typography.Paragraph>
            </div>

            <Alert
              type="info"
              showIcon
              message="调用方只需要员工 Key"
              description="示例里的 Authorization 是调用方访问网关使用的员工 Key。上游官方 API Key 已由网关按当前上游配置代为携带。"
            />
          </Space>
        )}
      </Modal>
    </Card>
  )
}
