发起对话

调用此接口发起一次对话，支持添加上下文和流式响应。​

会话、对话和消息的概念说明，可参考​基础概念。​

接口说明​

发起对话接口用于向指定智能体发起一次对话，支持在对话时添加对话的上下文消息，以便智能体基于历史消息做出合理的回复。开发者可以按需选择响应方式，即流式或非流式响应，响应方式决定了开发者获取智能体回复的方式。关于获取智能体回复的详细说明可参考​通过对话接口获取智能体回复。​

*   流式响应：智能体在生成回复的同时，将回复消息以数据流的形式逐条发送给客户端。处理结束后，服务端会返回一条完整的智能体回复。详细说明可参考​流式响应。​

*   非流式响应：无论对话是否处理完毕，立即发送响应消息。开发者可以通过接口​查看对话详情确认本次对话处理结束后，再调用​查看对话消息详情接口查看模型回复等完整响应内容。详细说明可参考​非流式响应。​

基础信息​

​

|     |     |
| --- | --- |
| 请求方式​ | POST​ |
| 请求地址​ | ​<br><br>Plain Text<br><br>复制<br><br>https://api.coze.cn/v3/chat​<br><br>​ |
| 权限​ | chat​<br><br>确保调用该接口使用的个人令牌开通了 chat 权限，详细信息参考​鉴权方式。​ |
| 接口说明​ | 调用此接口发起一次对话，支持添加上下文和流式响应。​ |

​

Header​

​

|     |     |     |
| --- | --- | --- |
| 参数​ | 取值​ | 说明​ |
| Authorization​ | Bearer $Access\_Token​ | 用于验证客户端身份的访问令牌。你可以在扣子平台中生成访问令牌，详细信息，参考​准备工作。​ |
| Content-Type​ | application/json​ | 解释请求正文的方式。​ |

​

​

​

Query​

​

|     |     |     |     |
| --- | --- | --- | --- |
| 参数​ | 类型​ | 是否必选​ | 说明​ |
| conversation\_id​ | String​ | 可选​ | 标识对话发生在哪一次会话中。​<br><br>会话是智能体和用户之间的一段问答交互。一个会话包含一条或多条消息。对话是会话中对智能体的一次调用，智能体会将对话中产生的消息添加到会话中。​<br><br>*   可以使用已创建的会话，会话中已存在的消息将作为上下文传递给模型。创建会话的方式可参考​创建会话。​<br><br>*   对于一问一答等不需要区分 conversation 的场合可不传该参数，系统会自动生成一个会话。 ​<br><br>说明<br><br>一个会话中，只能有一个进行中的对话，否则调用此接口时会报错 4016。​<br><br>​ |

​

Body​

​

|     |     |     |     |
| --- | --- | --- | --- |
| 参数​ | 类型​ | 是否必选​ | 说明​ |
| bot\_id​ | String​ | 必选​ | 要进行会话聊天的智能体ID。​<br><br>进入智能体的 开发页面，开发页面 URL 中 bot 参数后的数字就是智能体ID。例如https://www.coze.cn/space/341\*\*\*\*/bot/73428668\*\*\*\*\*，智能体ID 为73428668\*\*\*\*\*。​<br><br>说明<br><br>确保当前使用的访问密钥已被授予智能体所属空间的 chat 权限。​<br><br>​ |
| user\_id​ | String​ | 必选​ | 标识当前与智能体的用户，由使用方自行定义、生成与维护。user\_id 用于标识对话中的不同用户，不同的 user\_id，其对话的上下文消息、数据库等对话记忆数据互相隔离。如果不需要用户数据隔离，可将此参数固定为一个任意字符串，例如 123，abc 等。​<br><br>说明<br><br>出于数据隐私及信息安全等方面的考虑，不建议使用业务系统中定义的用户 ID。​<br><br>​ |
| additional\_messages​<br><br>​<br><br>​ | Array of object​<br><br>​ | 可选​<br><br>​ | 对话的附加信息。你可以通过此字段传入历史消息和本次对话中用户的问题。数组长度限制为 100，即最多传入 100 条消息。​<br><br>*   若未设置 additional\_messages，智能体收到的消息只有会话中已有的消息内容，其中最后一条作为本次对话的用户输入，其他内容均为本次对话的上下文。​<br><br>*   若设置了 additional\_messages，智能体收到的消息包括会话中已有的消息和 additional\_messages 中添加的消息，其中 additional\_messages 最后一条消息会作为本次对话的用户输入，其他内容均为本次对话的上下文。​<br><br>消息结构可参考​EnterMessage Object，具体示例可参考​携带上下文。​<br><br>说明<br><br>*   会话或 additional\_messages 中最后一条消息应为 role=user 的记录，以免影响模型效果。​<br><br>*   如果本次对话未指定会话或指定的会话中无消息时，必须通过此参数传入智能体用户的问题。​<br><br>​ |
| stream​<br><br>​ | Boolean​<br><br>​ | 可选​<br><br>​ | 是否启用流式返回。​<br><br>*   true：采用流式响应。 “流式响应”将模型的实时响应提供给客户端，类似打字机效果。你可以实时获取服务端返回的对话、消息事件，并在客户端中同步处理、实时展示，也可以直接在 completed 事件中获取智能体最终的回复。​<br><br>*   false：（默认）采用非流式响应。 “非流式响应”是指响应中仅包含本次对话的状态等元数据。此时应同时开启 auto\_save\_history，在本次对话处理结束后再查看模型回复等完整响应内容。可以参考以下业务流程：​<br><br>a.<br><br>调用发起对话接口，并设置 stream = false，auto\_save\_history=true，表示使用非流式响应，并记录历史消息。​<br><br>*   你需要记录会话的 Conversation ID 和 Chat ID，用于后续查看详细信息。​<br><br>b.<br><br>定期轮询​查看对话详情接口，建议每次间隔 1 秒以上，直到会话状态流转为终态，即 status 为 completed、required\_action、canceled 或 failed。​<br><br>c.<br><br>调用​查看对话消息详情接口，查询大模型生成的最终结果。​ |
| custom\_variables​ | Map<String, String>​ | 可选​ | 智能体提示词中定义的变量。在智能体 prompt 中设置变量 {{key}} 后，可以通过该参数传入变量值，同时支持 Jinja2 语法。详细说明可参考​变量示例。​<br><br>说明<br><br>*   仅适用于智能体提示词中定义的变量，不支持用于智能体的变量，或者传入到工作流中。​<br><br>*   变量名只支持英文字母和下划线。​<br><br>​ |
| auto\_save\_history​<br><br>​ | Boolean​<br><br>​ | 可选​ | 是否保存本次对话记录。​<br><br>*   true：（默认）会话中保存本次对话记录，包括 additional\_messages 中指定的所有消息、本次对话的模型回复结果、模型执行中间结果。​<br><br>*   false：会话中不保存本次对话记录，后续也无法通过任何方式查看本次对话信息、消息详情。在同一个会话中再次发起对话时，本次会话也不会作为上下文传递给模型。​<br><br>说明<br><br>非流式响应下（stream=false），此参数必须设置为 true，即保存本次对话记录，否则无法查看对话状态和模型回复。​<br><br>​ |
| meta\_data​ | Map​ | 可选​ | 附加信息，通常用于封装一些业务相关的字段。查看对话消息详情时，系统会透传此附加信息。​<br><br>自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。​ |
| extra\_params​ | Map<String, String>​ | 可选​ | 附加参数，通常用于特殊场景下指定一些必要参数供模型判断，例如指定经纬度，并询问智能体此位置的天气。​<br><br>自定义键值对格式，其中键（key）仅支持设置为：​<br><br>*   latitude：纬度，此时值（Value）为纬度值，例如 39.9800718。​<br><br>*   longitude：经度，此时值（Value）为经度值，例如 116.309314。​ |
| shortcut\_command​ | Object​ | 可选​ | 快捷指令信息。你可以通过此参数指定此次对话执行的快捷指令，必须是智能体已绑定的快捷指令。​<br><br>消息结构可参考 ShortcutCommandDetail Object。​<br><br>说明<br><br>调用快捷指令，会自动根据快捷指令配置信息生成本次对话中的用户问题，并放入 additional\_messages 最后一条消息作为本次对话的用户输入。​<br><br>​ |

​

EnterMessage Object​

​

|     |     |     |     |
| --- | --- | --- | --- |
| 参数​ | 类型​ | 是否必选​ | 说明​ |
| role​ | String​ | 必选​ | 发送这条消息的实体。取值：​<br><br>*   user：代表该条消息内容是用户发送的。​<br><br>*   assistant：代表该条消息内容是智能体发送的。​ |
| type​<br><br>​ | String​ | 可选​<br><br>​ | 消息类型。默认为 question。​<br><br>*   question：用户输入内容。​<br><br>*   answer：智能体返回给用户的消息内容，支持增量返回。如果工作流绑定了消息节点，可能会存在多 answer 场景，此时可以用流式返回的结束标志来判断所有 answer 完成。​<br><br>*   function\_call：智能体对话过程中调用函数（function call）的中间结果。 ​<br><br>*   tool\_response：调用工具 （function call）后返回的结果。​<br><br>*   follow\_up：如果在 智能体上配置打开了用户问题建议开关，则会返回推荐问题相关的回复内容。不支持在请求中作为入参。​<br><br>*   verbose：多 answer 场景下，服务端会返回一个 verbose 包，对应的 content 为 JSON 格式，content.msg\_type =generate\_answer\_finish 代表全部 answer 回复完成。不支持在请求中作为入参。​<br><br>说明<br><br>仅发起会话（v3）接口支持将此参数作为入参，且：​<br><br>*   如果 autoSaveHistory=true，type 支持设置为 question 或 answer。​<br><br>*   如果 autoSaveHistory=false，type 支持设置为 question、answer、function\_call、tool\_output/tool\_response。​<br><br>其中，type=question 只能和 role=user 对应，即仅用户角色可以且只能发起 question 类型的消息。详细说明可参考​消息 type 说明。​<br><br>​ |
| content​ | String​ | 可选​ | 消息的内容，支持纯文本、多模态（文本、图片、文件混合输入）、卡片等多种类型的内容。​<br><br>*   content\_type 为 object\_string 时，content 为 object\_string object 数组序列化之后的 JSON String，详细说明可参考 object\_string object。​<br><br>*   当 content\_type = text 时，content 为普通文本，例如 "content" :"Hello!"。​ |
| content\_type​ | String​ | 可选​ | 消息内容的类型，支持设置为：​<br><br>*   text：文本。​<br><br>*   object\_string：多模态内容，即文本和文件的组合、文本和图片的组合。​<br><br>*   card：卡片。此枚举值仅在接口响应中出现，不支持作为入参。​<br><br>说明<br><br>content 不为空时，此参数为必选。​<br><br>​ |
| meta\_data​ | Map ​ | 可选​ | 创建消息时的附加消息，获取消息时也会返回此附加消息。​<br><br>自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。​ |

​

object\_string object​

​

|     |     |     |     |
| --- | --- | --- | --- |
| 参数​ | 类型​ | 是否必选​ | 说明​ |
| type​ | String​ | 必选​ | 多模态消息内容类型，支持设置为：​<br><br>*   text：文本类型。​<br><br>*   file：文件类型。​<br><br>*   image：图片类型。​<br><br>*   audio：音频类型。​ |
| text​ | String​ | 可选​ | 文本内容。​ |
| file\_id​ | String​ | 可选​ | 文件、图片、音频内容的 ID。​<br><br>说明<br><br>*   必须是当前账号上传的文件 ID，上传方式可参考​上传文件。​<br><br>*   在 type 为 file、image 或 audio 时，file\_id 和 file\_url 应至少指定一个。​<br><br>​ |
| file\_url​ | String​ | 可选​ | 文件、图片或语音文件的在线地址。必须是可公共访问的有效地址。​<br><br>在 type 为 file、image 或 audio 时，file\_id 和 file\_url 应至少指定一个。​ |

​

说明

*   一个数组中只能有一个 Text 类型消息，但可以有多个 file、image 类型的消息。​

*   object\_string 中，text 需要与 file 或 image 一起使用。如果消息内容只有 text 类型，建议直接指定 content\_type: text，不使用 object\_string。​

*   支持发送纯图片或纯文件消息，但此类消息前后应同时传入一条文本消息作为用户 Query。例如 "content": "\[{\\"type\\":\\"image\\",\\"file\_id\\":\\"{{file\_id\_1}}\\"}\]" 为一条纯图片消息，且前后无文本消息，此时接口会报 4000 参数错误。​

​

例如，以下数组是一个完整的多模态内容：​

序列化前：​

​

JSON

复制

\[​

{​

"type": "text",​

"text": "你好我有一个帽衫，我想问问它好看么，你帮我看看"​

}, {​

"type": "image",​

"file\_id": "{{file\_id\_1}}"​

}, {​

"type": "file",​

"file\_id": "{{file\_id\_2}}"​

},​

{​

"type": "file",​

"file\_url": "{{file\_url\_1}}"​

}​

\]​

​

序列化后：​

​

JSON

复制

"\[{\\"type\\":\\"text\\",\\"text\\":\\"你好我有一个帽衫，我想问问它好看么，你帮我看看\\"},{\\"type\\":\\"image\\",\\"file\_id\\":\\"{{file\_id\_1}}\\"},{\\"type\\":\\"file\\",\\"file\_id\\":\\"{{file\_id\_2}}\\"},{\\"type\\":\\"file\\",\\"file\_url\\":\\"{{file\_url\_1}}\\"}\]"​

​

​

​

消息结构示例：​

文本消息

多模态消息

文本消息的 content\_type 为 text，消息结构示例如下。​

​

JSON

复制

{​

"role": "user",​

"content": "搜几个最新的军事新闻",​

"content\_type": "text"​

}​

​

​

ShortcutCommandDetail Object​

​

|     |     |     |     |
| --- | --- | --- | --- |
| 参数​ | 类型​ | 是否必选​ | 说明​ |
| command\_id​ | String​ | 必选​ | 对话要执行的快捷指令 ID，必须是智能体已绑定的快捷指令。​<br><br>你可以通过​获取智能体配置接口中的​ShortcutCommandInfo查看快捷指令 ID。​ |
| parameters​ | Map<String, String>​ | 可选​ | 用户输入的快捷指令组件参数信息。​<br><br>自定义键值对，其中键（key）为快捷指令组件的名称，值（value）为组件对应的用户输入，为 object\_string object 数组序列化之后的 JSON String，详细说明可参考 object\_string object。​ |

​

返回结果​

此接口通过请求 Body 参数 stream 为 true 或 false 来指定 Response 为流式或非流式响应。你可以根据以下步骤判断当前业务场景适合的响应模式。​

流程图

No

是否需要打字机效果  
显示Response

是否需要  
同步处理 Response

流式响应  
stream=true

非流式响应  
stream=false

是否需要  
即时查看 Bot 回复

YES

YES

YES

No

No

​

流式响应​

在流式响应中，服务端不会一次性发送所有数据，而是以数据流的形式逐条发送数据给客户端，数据流中包含对话过程中触发的各种事件（event），直至处理完毕或处理中断。处理结束后，服务端会通过 conversation.message.completed 事件返回拼接后完整的模型回复信息。各个事件的说明可参考流式响应事件。​

流式响应允许客户端在接收到完整的数据流之前就开始处理数据，例如在对话界面实时展示智能体的回复内容，减少客户端等待模型完整回复的时间。​

流式响应的整体流程如下：​

流式响应流程

流式响应示例

​

JSON

复制

\# chat - 开始​

event: conversation.chat.created​

// 在 chat 事件里，data 字段中的 id 为 Chat ID，即会话 ID。​

data: {"id": "123", "conversation\_id":"123", "bot\_id":"222", "created\_at":1710348675,compleated\_at:null, "last\_error": null, "meta\_data": {}, "status": "created","usage":null}​

​

\# chat - 处理中​

event: conversation.chat.in\_progress​

data: {"id": "123", "conversation\_id":"123", "bot\_id":"222", "created\_at":1710348675, compleated\_at: null, "last\_error": null,"meta\_data": {}, "status": "in\_progress","usage":null}​

​

\# MESSAGE - 知识库召回​

event: conversation.message.completed​

data: {"id": "msg\_001", "role":"assistant","type":"knowledge","content":"---\\nrecall slice 1:xxxxxxx\\n","content\_type":"text","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

​

\# MESSAGE - function\_call​

event: conversation.message.completed​

data: {"id": "msg\_002", "role":"assistant","type":"function\_call","content":"{\\"name\\":\\"toutiaosousuo-search\\",\\"arguments\\":{\\"cursor\\":0,\\"input\_query\\":\\"今天的体育新闻\\",\\"plugin\_id\\":7281192623887548473,\\"api\_id\\":7288907006982012986,\\"plugin\_type\\":1","content\_type":"text","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

​

\# MESSAGE - toolOutput​

event: conversation.message.completed​

data: {"id": "msg\_003", "role":"assistant","type":"tool\_output","content":"........","content\_type":"card","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

​

\# MESSAGE - answer is card​

event: conversation.message.completed​

data: {"id": "msg\_004", "role":"assistant","type":"answer","content":"{{card\_json}}","content\_type":"card","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

​

\# MESSAGE - answer is normal text​

event: conversation.message.delta​

data:{"id": "msg\_005", "role":"assistant","type":"answer","content":"以下","content\_type":"text","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

​

event: conversation.message.delta​

data:{"id": "msg\_005", "role":"assistant","type":"answer","content":"是","content\_type":"text","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

​

...... {{ N 个 delta 消息包}} ......​

​

event: conversation.message.completed​

data:{"id": "msg\_005", "role":"assistant","type":"answer","content":"{{msg\_005 完整的结果。即之前所有 msg\_005 delta 内容拼接的结果}}","content\_type":"text","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

​

​

\# MESSAGE - 多 answer 的情况,会继续有 message.delta​

event: conversation.message.delta​

data:{"id": "msg\_006", "role":"assistant","type":"answer","content":"你好你好","content\_type":"text","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

​

...... {{ N 个 delta 消息包}} ......​

​

event: conversation.message.completed​

data:{"id": "msg\_006", "role":"assistant","type":"answer","content":"{{msg\_006 完整的结果。即之前所有 msg\_006 delta 内容拼接的结果}}","content\_type":"text","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

​

\# MESSAGE - Verbose （流式 plugin, 多 answer 结束，Multi-agent 跳转等场景）​

event: conversation.message.completed​

data:{"id": "msg\_007", "role":"assistant","type":"verbose","content":"{\\"msg\_type\\":\\"generate\_answer\_finish\\",\\"data\\":\\"\\"}","content\_type":"text","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

​

\# MESSAGE - suggestion​

event: conversation.message.completed​

data: {"id": "msg\_008", "role":"assistant","type":"follow\_up","content":"朗尼克的报价是否会成功？","content\_type":"text","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

event: conversation.message.completed​

data: {"id": "msg\_009", "role":"assistant","type":"follow\_up","content":"中国足球能否出现？","content\_type":"text","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

event: conversation.message.completed​

data: {"id": "msg\_010", "role":"assistant","type":"follow\_up","content":"羽毛球种子选手都有谁？","content\_type":"text","chat\_id": "123", "conversation\_id":"123", "bot\_id":"222"}​

​

\# chat - 完成​

event: conversation.chat.completed （chat完成）​

data: {"id": "123", "chat\_id": "123", "conversation\_id":"123", "bot\_id":"222", "created\_at":1710348675, compleated\_at:1710348675, "last\_error":null, "meta\_data": {}, "status": "compleated", "usage":{"token\_count":3397,"output\_tokens":1173,"input\_tokens":2224}}​

​

event: done （stream流结束）​

data: \[DONE\]​

​

\# chat - 失败​

event: conversation.chat.failed​

data: {​

"code":701231,​

"msg":"error"​

}​

​

​

返回的事件消息体结构如下：​

​

|     |     |     |
| --- | --- | --- |
| 参数​ | 类型​ | 说明​ |
| event​ | String​ | 当前流式返回的数据包事件。详细说明可参考 流式响应事件。​ |
| data​ | Object​ | 消息内容。其中，chat 事件和 message 事件的格式不同。​<br><br>*   chat 事件中，data 为 Chat Object。​<br><br>*   message、audio 事件中，data 为 Message Object。​ |

​

流式响应事件​

​

|     |     |
| --- | --- |
| 事件（event）名称​ | 说明​ |
| conversation.chat.created​ | 创建对话的事件，表示对话开始。​ |
| conversation.chat.in\_progress​ | 服务端正在处理对话。​ |
| conversation.message.delta​ | 增量消息，通常是 type=answer 时的增量消息。​ |
| conversation.audio.delta​ | 增量语音消息，通常是 type=answer 时的增量消息。​ |
| conversation.message.completed​ | message 已回复完成。此时流式包中带有所有 message.delta 的拼接结果，且每个消息均为 completed 状态。​ |
| conversation.chat.completed​ | 对话完成。​ |
| conversation.chat.failed​ | 此事件用于标识对话失败。​ |
| conversation.chat.requires\_action​ | 对话中断，需要使用方上报工具的执行结果。​ |
| error​ | 流式响应过程中的错误事件。关于 code 和 msg 的详细说明，可参考​错误码。​ |
| done​ | 本次会话的流式返回正常结束。​ |

​

非流式响应​

在非流式响应中，无论服务端是否处理完毕，立即发送响应消息。其中包括本次对话的 chat\_id、状态等元数据信息，但不包括模型处理的最终结果。​

非流式响应不需要维持长连接，在场景实现上更简单，但通常需要客户端主动查询对话状态和消息详情才能得到完整的数据。你可以通过接口​查看对话详情确认本次对话处理结束后，再调用​查看对话消息详情接口查看模型回复等完整响应内容。流程如下：​

​

​​![](data:image/svg+xml,%3csvg%20xmlns=%27http://www.w3.org/2000/svg%27%20version=%271.1%27%20width=%27282%27%20height=%27340%27/%3e)![](https://p9-arcosite.byteimg.com/https://p9-arcosite.byteimg.com/obj/tos-cn-i-goo7wpa0wc/79f9ce1cefc94f4daf20cad241382cdf~tplv-goo7wpa0wc-quality:q75.image)​

​

非流式响应的结构如下：​

​

|     |     |     |
| --- | --- | --- |
| 参数​ | 类型​ | 说明​ |
| data​ | Object​ | 本次对话的基本信息。详细说明可参考 Chat Object。​ |
| code​ | Integer​ | 状态码。​<br><br>0 代表调用成功。​ |
| msg​ | String​ | 状态信息。API 调用失败时可通过此字段查看详细错误信息。​ |

​

​

​

Chat Object​

​

|     |     |     |     |
| --- | --- | --- | --- |
| 参数​ | 类型​ | 是否可选​ | 说明​ |
| id​ | String​ | 必填​ | 对话 ID，即对话的唯一标识。​ |
| conversation\_id​ | String​ | 必填​ | 会话 ID，即会话的唯一标识。​ |
| bot\_id​ | String​ | 必填​ | 要进行会话聊天的智能体 ID。​ |
| created\_at​ | Integer​ | 选填​ | 对话创建的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。​ |
| completed\_at​ | Integer​ | 选填​ | 对话结束的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。​ |
| failed\_at​ | Integer​ | 选填​ | 对话失败的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。​ |
| meta\_data​ | Map<String,String>​ | 选填​ | 创建消息时的附加消息，用于传入使用方的自定义数据，获取消息时也会返回此附加消息。​<br><br>自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。​ |
| last\_error​<br><br>​ | Object​ | 选填​ | 对话运行异常时，此字段中返回详细的错误信息，包括：​<br><br>*   Code：错误码。Integer 类型。0 表示成功，其他值表示失败。​<br><br>*   Msg：错误信息。String 类型。​<br><br>说明<br><br>*   对话正常运行时，此字段返回 null。​<br><br>*   suggestion 失败不会被标记为运行异常，不计入 last\_error。​<br><br>​ |
| status​<br><br>​ | String​ | 必填​ | 对话的运行状态。取值为：​<br><br>*   created：对话已创建。​<br><br>*   in\_progress：智能体正在处理中。​<br><br>*   completed：智能体已完成处理，本次对话结束。​<br><br>*   failed：对话失败。​<br><br>*   requires\_action：对话中断，需要进一步处理。​<br><br>*   canceled：对话已取消。​ |
| required\_action​ | Object​ | 选填​ | 需要运行的信息详情。​ |
| » type​ | String​ | 选填​ | 额外操作的类型，枚举值为 submit\_tool\_outputs。​ |
| »submit\_tool\_outputs​ | Object​ | 选填​ | 需要提交的结果详情，通过提交接口上传，并可以继续聊天​ |
| »» tool\_calls​ | Array of Object​ | 选填​ | 具体上报信息详情。​ |
| »»» id​ | String​ | 选填​ | 上报运行结果的 ID。​ |
| »»» type​ | String​ | 选填​ | 工具类型，枚举值包括：​<br><br>*   function：待执行的方法，通常是端插件。触发端插件时会返回此枚举值。​<br><br>*   reply\_message：待回复的选项。触发工作流问答节点时会返回此枚举值。​ |
| »»» function​ | Object​ | 选填​ | 执行方法 function 的定义。​ |
| »»»» name​ | String​ | 选填​ | 方法名。​ |
| »»»» arguments​ | String​ | 选填​ | 方法参数。​ |
| usage​ | Object​ | 选填​ | Token 消耗的详细信息。实际的 Token 消耗以对话结束后返回的值为准。​ |
| »token\_count​ | Integer​ | 选填​ | 本次对话消耗的 Token 总数，包括 input 和 output 部分的消耗。​ |
| »output\_count​ | Integer​ | 选填​ | output 部分消耗的 Token 总数。​ |
| »input\_count​ | Integer​ | 选填​ | input 部分消耗的 Token 总数。​ |

​

Chat Object 的示例如下：​

状态正常的对话

需要使用方额外处理的对话

​

JSON

复制

{​

// 在 chat 事件里，data 字段中的 id 为 Chat ID，即会话 ID。​

"id": "737662389258662\*\*\*\*",​

"conversation\_id": "737554565555041\*\*\*\*",​

"bot\_id": "736661612448078\*\*\*\*",​

"completed\_at": 1717508113,​

"last\_error": {​

"code": 0,​

"msg": ""​

},​

"status": "completed",​

"usage": {​

"token\_count": 6644,​

"output\_count": 766,​

"input\_count": 5878​

}​

}​

​

​

*   ​

​

Message Object​

​

|     |     |     |
| --- | --- | --- |
| 参数​ | 类型​ | 说明​ |
| id​ | String​ | Message ID，即消息的唯一标识。​ |
| conversation\_id​ | String​ | 此消息所在的会话 ID。​ |
| bot\_id​ | String​ | 编写此消息的智能体ID。此参数仅在对话产生的消息中返回。​ |
| chat\_id​ | String​ | Chat ID。此参数仅在对话产生的消息中返回。​ |
| meta\_data​ | Map​ | 创建消息时的附加消息，获取消息时也会返回此附加消息。​ |
| role​ | String​ | 发送这条消息的实体。取值：​<br><br>*   user：代表该条消息内容是用户发送的。​<br><br>*   assistant：代表该条消息内容是智能体发送的。​ |
| content​ | String​<br><br>​ | 消息的内容，支持纯文本、多模态（文本、图片、文件混合输入）、卡片等多种类型的内容。​ |
| content\_type​ | String​ | 消息内容的类型，取值包括：​<br><br>*   text：文本。​<br><br>*   object\_string：多模态内容，即文本和文件的组合、文本和图片的组合。​<br><br>*   card：卡片。此枚举值仅在接口响应中出现，不支持作为入参。​<br><br>*   audio：音频。此枚举值仅在接口响应中出现，不支持作为入参。仅当输入有 audio 文件时，才会返回此类型。当 content\_type 为 audio 时，content 为 base64 后的音频数据。音频的编码根据输入的 audio 文件的不同而不同：​<br><br>*   输入为 wav 格式音频时，content 为采样率 24kHz，raw 16 bit, 1 channel, little-endian 的 pcm 音频片段 base64 后的字符串​<br><br>*   输入为 ogg\_opus 格式音频时，content 为采样率 48kHz，1 channel，10ms 帧长的 opus 格式音频片段base64 后的字符串​ |
| created\_at​ | Integer​ | 消息的创建时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。​ |
| updated\_at​ | Integer​ | 消息的更新时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。​ |
| type​ | String​ | 消息类型。​<br><br>*   question：用户输入内容。​<br><br>*   answer：智能体返回给用户的消息内容，支持增量返回。如果工作流绑定了 messge 节点，可能会存在多 answer 场景，此时可以用流式返回的结束标志来判断所有 answer 完成。​<br><br>*   function\_call：智能体对话过程中调用函数（function call）的中间结果。​<br><br>*   tool\_response：调用工具 （function call）后返回的结果。​<br><br>*   follow\_up：如果在智能体上配置打开了用户问题建议开关，则会返回推荐问题相关的回复内容。​<br><br>*   verbose：多 answer 场景下，服务端会返回一个 verbose 包，对应的 content 为 JSON 格式，content.msg\_type =generate\_answer\_finish 代表全部 answer 回复完成。​<br><br>说明<br><br>仅发起会话（v3）接口支持将此参数作为入参，且：​<br><br>*   如果 autoSaveHistory=true，type 支持设置为 question 或 answer。​<br><br>*   如果 autoSaveHistory=false，type 支持设置为 question、answer、function\_call、tool\_response。​<br><br>其中，type=question 只能和 role=user 对应，即仅用户角色可以且只能发起 question 类型的消息。详细说明可参考​消息 type 说明。​<br><br>​ |
| section\_id​ | String​ | 上下文片段 ID。每次清除上下文都会生成一个新的 section\_id。​ |
| reasoning\_content​ | String​ | DeepSeek-R1 模型的思维链（CoT）。模型会将复杂问题逐步分解为多个简单步骤，并按照这些步骤逐一推导出最终答案。​<br><br>该参数仅在使用 DeepSeek-R1 模型时才会返回。​ |

​

示例​

流式响应​

基础问答

图文问答

*   Request​

*   ​
    
    Shell
    
    复制
    
    curl --location --request POST 'https://api.coze.cn/v3/chat?conversation\_id=7374752000116113452' \\​
    
    \--header 'Authorization: Bearer pat\_OYDacMzM3WyOWV3Dtj2bHRMymzxP\*\*\*\*' \\​
    
    \--header 'Content-Type: application/json' \\​
    
    \--data-raw '{​
    
    "bot\_id": "734829333445931\*\*\*\*",​
    
    "user\_id": "123456789",​
    
    "stream": true,​
    
    "auto\_save\_history":true,​
    
    "additional\_messages":\[​
    
    {​
    
    "role":"user",​
    
    "content":"2024年10月1日是星期几",​
    
    "content\_type":"text"​
    
    }​
    
    \]​
    
    }'​
    
    ​

*   Response​

*   ​
    
    JSON
    
    复制
    
    event:conversation.chat.created​
    
    // 在 chat 事件里，data 字段中的 id 为 Chat ID，即会话 ID。​
    
    data:{"id":"7382159487131697202","conversation\_id":"7381473525342978089","bot\_id":"7379462189365198898","completed\_at":1718792949,"last\_error":{"code":0,"msg":""},"status":"created","usage":{"token\_count":0,"output\_count":0,"input\_count":0}}​
    
    ​
    
    event:conversation.chat.in\_progress​
    
    data:{"id":"7382159487131697202","conversation\_id":"7381473525342978089","bot\_id":"7379462189365198898","completed\_at":1718792949,"last\_error":{"code":0,"msg":""},"status":"in\_progress","usage":{"token\_count":0,"output\_count":0,"input\_count":0}}​
    
    ​
    
    event:conversation.message.delta​
    
    data:{"id":"7382159494123470858","conversation\_id":"7381473525342978089","bot\_id":"7379462189365198898","role":"assistant","type":"answer","content":"2","content\_type":"text","chat\_id":"7382159487131697202"}​
    
    ​
    
    event:conversation.message.delta​
    
    data:{"id":"7382159494123470858","conversation\_id":"7381473525342978089","bot\_id":"7379462189365198898","role":"assistant","type":"answer","content":"0","content\_type":"text","chat\_id":"7382159487131697202"}​
    
    ​
    
    //省略模型回复的部分中间事件event:conversation.message.delta​
    
    ......​
    
    ​
    
    event:conversation.message.delta​
    
    data:{"id":"7382159494123470858","conversation\_id":"7381473525342978089","bot\_id":"7379462189365198898","role":"assistant","type":"answer","content":"星期三","content\_type":"text","chat\_id":"7382159487131697202"}​
    
    ​
    
    event:conversation.message.delta​
    
    data:{"id":"7382159494123470858","conversation\_id":"7381473525342978089","bot\_id":"7379462189365198898","role":"assistant","type":"answer","content":"。","content\_type":"text","chat\_id":"7382159487131697202"}​
    
    ​
    
    event:conversation.message.completed​
    
    data:{"id":"7382159494123470858","conversation\_id":"7381473525342978089","bot\_id":"7379462189365198898","role":"assistant","type":"answer","content":"2024 年 10 月 1 日是星期三。","content\_type":"text","chat\_id":"7382159487131697202"}​
    
    ​
    
    event:conversation.message.completed​
    
    data:{"id":"7382159494123552778","conversation\_id":"7381473525342978089","bot\_id":"7379462189365198898","role":"assistant","type":"verbose","content":"{\\"msg\_type\\":\\"generate\_answer\_finish\\",\\"data\\":\\"\\",\\"from\_module\\":null,\\"from\_unit\\":null}","content\_type":"text","chat\_id":"7382159487131697202"}​
    
    ​
    
    event:conversation.chat.completed​
    
    data:{"id":"7382159487131697202","conversation\_id":"7381473525342978089","bot\_id":"7379462189365198898","completed\_at":1718792949,"last\_error":{"code":0,"msg":""},"status":"completed","usage":{"token\_count":633,"output\_count":19,"input\_count":614}}​
    
    ​
    
    event:done​
    
    data:"\[DONE\]"​
    
    ​

​

非流式响应​

基础问答

图文问答

快捷指令

*   Request​

*   ​
    
    Shell
    
    复制
    
    curl --location --request POST 'https://api.coze.cn/v3/chat?conversation\_id=737475200011611\*\*\*\*' \\​
    
    \--header 'Authorization: Bearer pat\_OYDacMzM3WyOWV3Dtj2bHRMymzxP\*\*\*\*' \\​
    
    \--header 'Content-Type: application/json' \\​
    
    \--data-raw '{​
    
    "bot\_id": "734829333445931\*\*\*\*",​
    
    "user\_id": "123456789",​
    
    "stream": false,​
    
    "auto\_save\_history":true,​
    
    "additional\_messages":\[​
    
    {​
    
    "role":"user",​
    
    "content":"今天杭州天气如何",​
    
    "content\_type":"text"​
    
    }​
    
    \]​
    
    }'​
    
    ​

*   Response​

*   ​
    
    JSON
    
    复制
    
    {​
    
    "data":{​
    
    // data 字段中的 id 为 Chat ID，即会话 ID。​
    
    "id": "123",​
    
    "conversation\_id": "123456",​
    
    "bot\_id": "222",​
    
    "created\_at": 1710348675,​
    
    "completed\_at": 1710348675,​
    
    "last\_error": null,​
    
    "meta\_data": {},​
    
    "status": "completed",​
    
    "usage": {​
    
    "token\_count": 3397,​
    
    "output\_count": 1173,​
    
    "input\_count": 2224​
    
    }​
    
    },​
    
    "code":0,​
    
    "msg":""​
    
    }​
    
    ​

​

Prompt 变量​

例如在智能体的 Prompt 中定义了一个 {{bot\_name}} 的变量，在调用接口时，可以通过 custom\_variables 参数传入变量值。​

智能体Prompt 配置示例：​

​

​​![](data:image/svg+xml,%3csvg%20xmlns=%27http://www.w3.org/2000/svg%27%20version=%271.1%27%20width=%27212%27%20height=%27195%27/%3e)![](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/23fa1031548148bd969b3462f5b4e436~tplv-goo7wpa0wc-quality:q75.image)​

​

API 调用示例：​

​

​​![](data:image/svg+xml,%3csvg%20xmlns=%27http://www.w3.org/2000/svg%27%20version=%271.1%27%20width=%27347%27%20height=%27224%27/%3e)![](https://p9-arcosite.byteimg.com/https://p9-arcosite.byteimg.com/obj/tos-cn-i-goo7wpa0wc/9097bfa34a06460cabf3682f815e54a3~tplv-goo7wpa0wc-quality:q75.image)​

​

​

扣子也支持 Jinja2 语法。在下面这个模板中，prompt1 将在 key 变量存在时使用，而 prompt2 将在 key 变量不存在时使用。通过在 custom\_variables 中传递 key 的值，你可以控制智能体的响应。​

​

Python

复制

{% if key -%}​

prompt1​

{%- else %}​

prompt2​

{% endif %}​

​

智能体Prompt 配置示例：​

​

​​![](data:image/svg+xml,%3csvg%20xmlns=%27http://www.w3.org/2000/svg%27%20version=%271.1%27%20width=%27346%27%20height=%27218%27/%3e)![](https://p9-arcosite.byteimg.com/https://p9-arcosite.byteimg.com/obj/tos-cn-i-goo7wpa0wc/a137259f46954467879145d31999e4ee~tplv-goo7wpa0wc-quality:q75.image)​

​

API 调用示例：​

​

​​![](data:image/svg+xml,%3csvg%20xmlns=%27http://www.w3.org/2000/svg%27%20version=%271.1%27%20width=%27354%27%20height=%27239%27/%3e)![](https://p9-arcosite.byteimg.com/https://p9-arcosite.byteimg.com/obj/tos-cn-i-goo7wpa0wc/a2e07f36ca9943949889200bc51de7ec~tplv-goo7wpa0wc-quality:q75.image)​

​

​

携带上下文​

你可以在发起对话时把多条消息作为上下文一起上传，模型会参考上下文消息，对用户 Query 进行针对性回复。在发起对话时，扣子会将以下内容作为上下文传递给模型。​

*   会话中的消息：调用发起对话接口时，如果指定了会话 ID，会话中已有的消息会作为上下文传递给模型。​

*   additional\_messages 中的消息：如果 additional\_messages 中有多条消息，则最后一条会作为本次用户 Query，其他消息为上下文。​

扣子推荐你通过以下方式在对话中指定上下文：​

​

|     |     |
| --- | --- |
| 方式​ | 说明​ |
| 方式一：通过会话传递历史消息，通过 additional\_messages 指定用户 Query​ | 适用于在已有会话中再次发起对话的场景，会话中通常已经存在部分历史消息，开发者也可以手动插入一些消息作为上下文。​ |
| 方式二：通过 additional\_messages 指定历史消息和用户 Query​ | 此方式无需提前创建会话，通过发起对话一个接口即可完成一次携带上下文的对话，更适用于一问一答的场景，使用方式更加简便。​ |

​

以方式一为例，在对话中携带上下文的操作步骤如下：​

1.

准备上下文消息。​

*   说明
    
    准备上下文消息时应注意：​
    
    *   应包含用户询问和模型返回两部分消息数据。详情可参考返回参数内容中 Message 消息结构的具体说明。​
    
    *   上下文消息列表按时间递增排序，即最近一条 message 在列表的最后一位。​
    
    *   只需传入用户输入内容及模型返回内容即可，即 role=user 和 role=assistant; type=answer。​
    
    ​

*   以下消息列表是一个完整的上下文消息。其中：​

*   第 2 行是用户传入的历史消息​

*   第 4 行是模型返回的历史消息​

*   ​
    
    JSON
    
    复制
    
    \[​
    
    { "role": "user", "content\_type":"text", "content": "你可以读懂图片中的内容吗" }​
    
    ​
    
    {"role":"assistant","type":"answer","content":"没问题！你想查看什么图片呢？"，"content\_type":"text"}​
    
    \]​
    
    ​

2.

调用​创建会话接口创建一个会话，其中包含以上两条消息，并记录会话 ID。​

*   请求示例如下：​

*   ​
    
    Shell
    
    复制
    
    curl --location --request POST 'https://api.coze.cn/v1/conversation/create' \\​
    
    \--header 'Authorization: Bearer pat\_OYDacMzM3WyOWV3Dtj2bHRMymzxP\*\*\*\*' \\​
    
    \--header 'Content-Type: application/json' \\​
    
    \--data-raw '{​
    
    "meta\_data": {​
    
    "uuid": "newid1234"​
    
    },​
    
    "messages": \[​
    
    {​
    
    "role": "user",​
    
    "content":"你可以读懂图片中的内容吗",​
    
    "content\_type":"text"​
    
    },​
    
    {​
    
    "role": "assistant",​
    
    "type":"answer",​
    
    "content": "没问题！你想查看什么图片呢？",​
    
    "content\_type":"text"​
    
    }​
    
    \]​
    
    }'​
    
    ​

3.

调用发起对话（V3）接口，并指定会话 ID。​

*   在对话中可以通过 additional\_messages 增加本次对话的 query。这些消息会和对话中已有的消息一起作为上下文被传递给大模型。​

*   ​
    
    Shell
    
    复制
    
    curl --location --request POST 'https://api.coze.cn/v3/chat?conversation\_id=737363834493434\*\*\*\*' \\​
    
    \--header 'Authorization: Bearer pat\_OYDacMzM3WyOWV3Dtj2bHRMymzxP\*\*\*\*' \\​
    
    \--header 'Content-Type: application/json' \\​
    
    \--data-raw '{​
    
    "bot\_id": "734829333445931\*\*\*\*",​
    
    "user\_id": "123456789",​
    
    "stream": false,​
    
    "auto\_save\_history":true,​
    
    "additional\_messages":\[​
    
    {​
    
    "role":"user",​
    
    "content":"\[{\\"type\\":\\"image\\",\\"file\_url\\":\\"https://gimg2.baidu.com/image\_search/src=http%3A%2F%2Fci.xiaohongshu.com%2Fe7368218-8a64-bda3-56ad-5672b2a113b2%3FimageView2%2F2%2Fw%2F1080%2Fformat%2Fjpg&refer=http%3A%2F%2Fci.xiaohongshu.com&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=auto?sec=1720005307&t=1acd734e6e8937c2d77d625bcdb0dc57\\"},{\\"type\\":\\"text\\",\\"text\\":\\"这张可以吗\\"}\]",​
    
    "content\_type":"object\_string"​
    
    }​
    
    \]​
    
    }'​
    
    ​

4.

调用接口​查看对话消息详情查看模型回复。​

*   你可以从智能体的回复中看出这一次会话是符合上下文语境的。响应信息如下：​

*   ​
    
    JSON
    
    复制
    
    {​
    
    "code": 0,​
    
    "data": \[​
    
    {​
    
    "bot\_id": "737946218936519\*\*\*\*",​
    
    "content": "{\\"name\\":\\"tupianlijie-imgUnderstand\\",\\"arguments\\":{\\"text\\":\\"图中是什么内容\\",\\"url\\":\\"https://lf-bot-studio-plugin-resource.coze.cn/obj/bot-studio-platform-plugin-tos/artist/image/4ca71a5f55d54efc95ed9c06e019ff4b.png\\"},\\"plugin\_id\\":7379227414322217010,\\"api\_id\\":7379227414322233394,\\"plugin\_type\\":1,\\"thought\\":\\"需求为识别图中（https://lf-bot-studio-plugin-resource.coze.cn/obj/bot-studio-platform-plugin-tos/artist/image/4ca71a5f55d54efc95ed9c06e019ff4b.png）的内容，需要调用tupianlijie-imgUnderstand进行识别\\"}",​
    
    "content\_type": "text",​
    
    "conversation\_id": "738147352534297\*\*\*\*",​
    
    "id": "7381473945440239668",​
    
    "role": "assistant",​
    
    "type": "function\_call"​
    
    },​
    
    {​
    
    "bot\_id": "7379462189365198898",​
    
    "content": "{\\"content\_type\\":1,\\"response\_for\_model\\":\\"图中展示的是一片茂密的树林。\\",\\"type\_for\_model\\":1}",​
    
    "content\_type": "text",​
    
    "conversation\_id": "738147352534297\*\*\*\*",​
    
    "id": "7381473964905807872",​
    
    "role": "assistant",​
    
    "type": "tool\_response"​
    
    },​
    
    {​
    
    "bot\_id": "7379462189365198898",​
    
    "content": "{\\"msg\_type\\":\\"generate\_answer\_finish\\",\\"data\\":\\"\\",\\"from\_module\\":null,\\"from\_unit\\":null}",​
    
    "content\_type": "text",​
    
    "conversation\_id": "738147352534297\*\*\*\*",​
    
    "id": "7381473964905906176",​
    
    "role": "assistant",​
    
    "type": "verbose"​
    
    },​
    
    {​
    
    "bot\_id": "7379462189365198898",​
    
    "content": "这幅图展示的是一片茂密的树林。",​
    
    "content\_type": "text",​
    
    "conversation\_id": "738147352534297\*\*\*\*",​
    
    "id": "7381473945440223284",​
    
    "role": "assistant",​
    
    "type": "answer"​
    
    }​
    
    \],​
    
    "msg": ""​
    
    }​
    
    ​

​