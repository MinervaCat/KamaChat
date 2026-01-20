// src/store/messageStore.js
import { reactive, computed, watch, toRefs } from 'vue'
import { cloneDeep } from 'lodash-es'// 可选，用于深拷贝
import store from "@/store/index";
import axios from "axios";

class MessageStore {
    constructor() {
        // 主数据结构
        this.state = reactive({
            // 会话索引：key=sessionId, value=会话详情
            sessions: new Map(),

            // 消息索引：key=messageId, value=消息对象
            messages: new Map(),

            // 会话-消息关联：key=sessionId, value=Set<messageId>
            sessionMessages: new Map(),

            // 用户索引：key=userId, value=用户信息
            users: new Map(),

            // 当前活动会话
            activeSessionId: null,

            // 未读消息计数：key=sessionId, value=未读数
            unreadCounts: new Map(),

            // 会话排序缓存（按最后消息时间排序）
            sortedSessionIds: [],

            // 快速搜索索引（可选，用于全文搜索）
            searchIndex: new Map(), // key=关键词, value=Set<messageId>

            msgSeq: 0
        })

        // 初始化一些工具函数
        this.init()
    }

    init() {
        // 可以添加一些初始化逻辑
    }
    fetchAllMessage() {
        console.log("开始获取所有聊天")
        try {
            const req = {
                user_id: store.state.userInfo.user_id,
                seq: this.state.msgSeq,
            };
            console.log(req);
            const rsp = axios.post(
                store.state.backendUrl + "/message/getMessageAfterSeq",
                req
            );
            if (rsp.data.data) {
                for (let i = 0; i < rsp.data.data.length; i++) {
                    if (!this.state.sessionMessages.has(rsp.data.data[i].conversation_id)) {
                        this.state.sessionMessages.set(rsp.data.data[i].conversation_id, new Set())
                    }
                    this.state.sessionMessages.get(rsp.data.data[i].conversation_id).add(rsp.data.data[i])
                }
            }
            console.log(rsp);
        } catch (error) {
            console.error(error);
        }
    }
    // ============ 会话管理 ============

    /**
     * 添加或更新会话
     * @param {Object} session - 会话对象
     * @param {string} session.id - 会话ID
     * @param {string} session.type - 会话类型: 'private' | 'group'
     * @param {string} session.name - 会话名称
     * @param {string} session.avatar - 会话头像
     * @param {Array} session.members - 成员列表
     * @param {number} session.lastMessageTime - 最后消息时间戳
     */
    upsertSession(session) {
        const sessionId = session.id

        // 如果会话已存在，合并更新
        if (this.state.sessions.has(sessionId)) {
            const existing = this.state.sessions.get(sessionId)
            this.state.sessions.set(sessionId, {
                ...existing,
                ...session,
                updatedAt: Date.now()
            })
        } else {
            // 新会话
            this.state.sessions.set(sessionId, {
                ...session,
                createdAt: Date.now(),
                updatedAt: Date.now()
            })

            // 初始化关联数据结构
            this.state.sessionMessages.set(sessionId, new Set())
            this.state.unreadCounts.set(sessionId, 0)
        }

        // 更新排序缓存
        this.updateSessionSorting(sessionId)

        return sessionId
    }

    /**
     * 获取会话信息
     */
    getSession(sessionId) {
        return this.state.sessions.get(sessionId) || null
    }

    /**
     * 获取所有会话（按最后消息时间排序）
     */
    getAllSessions() {
        return this.state.sortedSessionIds
            .map(id => this.getSession(id))
            .filter(Boolean)
    }

    /**
     * 更新会话排序
     */
    updateSessionSorting(sessionId) {
        const session = this.getSession(sessionId)
        if (!session) return

        // 从当前排序中移除
        const index = this.state.sortedSessionIds.indexOf(sessionId)
        if (index > -1) {
            this.state.sortedSessionIds.splice(index, 1)
        }

        // 插入到正确位置（按最后消息时间倒序）
        const lastMsgTime = session.lastMessageTime || 0
        let insertIndex = 0

        for (; insertIndex < this.state.sortedSessionIds.length; insertIndex++) {
            const otherSession = this.getSession(this.state.sortedSessionIds[insertIndex])
            if (otherSession && (otherSession.lastMessageTime || 0) < lastMsgTime) {
                break
            }
        }

        this.state.sortedSessionIds.splice(insertIndex, 0, sessionId)
    }

    // ============ 消息管理 ============

    /**
     * 添加消息
     * @param {Object} message - 消息对象
     * @param {string} message.id - 消息ID
     * @param {string} message.sessionId - 所属会话ID
     * @param {string} message.senderId - 发送者ID
     * @param {string} message.content - 消息内容
     * @param {number} message.timestamp - 时间戳
     * @param {string} message.type - 消息类型: 'text' | 'image' | 'file'
     * @param {string} message.status - 状态: 'sending' | 'sent' | 'read' | 'failed'
     */
    addMessage(message) {
        const { id, sessionId } = message

        // 1. 保存到消息索引
        this.state.messages.set(id, {
            ...message,
            _indexedAt: Date.now()
        })

        // 2. 添加到会话关联
        if (!this.state.sessionMessages.has(sessionId)) {
            this.state.sessionMessages.set(sessionId, new Set())
        }
        this.state.sessionMessages.get(sessionId).add(id)

        // 3. 更新会话的最后消息时间
        const session = this.getSession(sessionId)
        if (session) {
            session.lastMessageTime = message.timestamp
            session.lastMessage = message.content
            session.lastMessageId = id
            this.updateSessionSorting(sessionId)
        }

        // 4. 如果当前用户不是发送者，增加未读数
        if (message.senderId !== this.currentUserId && message.status !== 'read') {
            this.incrementUnreadCount(sessionId)
        }

        // 5. 建立搜索索引（可选，根据需求开启）
        // this.indexMessageForSearch(message)

        return id
    }

    /**
     * 批量添加消息（用于初始化加载）
     */
    addMessages(messages) {
        const results = {
            added: 0,
            skipped: 0,
            sessionIds: new Set()
        }

        // 按时间戳排序（确保按时间顺序添加）
        const sortedMessages = [...messages].sort((a, b) => a.timestamp - b.timestamp)

        sortedMessages.forEach(msg => {
            // 检查是否已存在
            if (this.state.messages.has(msg.id)) {
                results.skipped++
                return
            }

            this.addMessage(msg)
            results.added++
            results.sessionIds.add(msg.sessionId)
        })

        // 更新每个会话的排序
        results.sessionIds.forEach(sessionId => {
            this.updateSessionSorting(sessionId)
        })

        return results
    }

    /**
     * 获取消息
     */
    getMessage(messageId) {
        return this.state.messages.get(messageId) || null
    }

    /**
     * 获取会话中的消息列表（支持分页）
     * @param {string} sessionId - 会话ID
     * @param {Object} options - 选项
     * @param {number} options.limit - 返回数量
     * @param {number} options.beforeTime - 在此时间之前
     * @param {number} options.afterTime - 在此时间之后
     * @param {string} options.direction - 方向: 'older' | 'newer'
     */
    getSessionMsg(sessionId) {
        return this.state.sessionMessages.get(sessionId) || null
    }

    getSessionMessages(sessionId, options = {}) {
        const {
            limit = 50,
            beforeTime = Date.now(),
            afterTime = 0,
            direction = 'older'
        } = options

        // 获取该会话的所有消息ID
        const messageIds = this.state.sessionMessages.get(sessionId)
        if (!messageIds) return []

        // 转换为消息数组并过滤
        let messages = Array.from(messageIds)
            .map(id => this.getMessage(id))
            .filter(msg => {
                if (!msg) return false
                if (msg.timestamp >= beforeTime) return false
                if (msg.timestamp <= afterTime) return false
                return true
            })
            .sort((a, b) => b.timestamp - a.timestamp) // 时间倒序

        // 根据方向调整
        if (direction === 'older') {
            // 获取更旧的消息（时间戳更小）
            messages = messages.slice(0, limit)
        } else {
            // 获取更新的消息（时间戳更大）
            messages = messages.slice(-limit)
        }

        // 按时间正序返回（最旧的在前）
        return messages.reverse()
    }

    /**
     * 更新消息状态
     */

    updateMessage(messageId, updates) {
        const message = this.getMessage(messageId)
        if (!message) return false

        // 合并更新
        Object.assign(message, updates, {
            _updatedAt: Date.now()
        })

        // 如果是标记为已读，更新未读数
        if (updates.status === 'read' && updates.senderId !== this.currentUserId) {
            this.decrementUnreadCount(message.sessionId)
        }

        return true
    }

    /**
     * 删除消息
     */
    deleteMessage(messageId) {
        const message = this.getMessage(messageId)
        if (!message) return false

        const { sessionId } = message

        // 1. 从消息索引中删除
        this.state.messages.delete(messageId)

        // 2. 从会话关联中删除
        const sessionMsgSet = this.state.sessionMessages.get(sessionId)
        if (sessionMsgSet) {
            sessionMsgSet.delete(messageId)

            // 如果会话没有消息了，更新最后消息
            if (sessionMsgSet.size === 0) {
                const session = this.getSession(sessionId)
                if (session) {
                    session.lastMessage = null
                    session.lastMessageId = null
                    this.updateSessionSorting(sessionId)
                }
            }
        }

        // 3. 从搜索索引中删除（如果启用了）
        // this.removeFromSearchIndex(messageId)

        return true
    }

    // ============ 用户管理 ============

    /**
     * 添加或更新用户信息
     */
    upsertUser(user) {
        const userId = user.id
        if (this.state.users.has(userId)) {
            const existing = this.state.users.get(userId)
            this.state.users.set(userId, { ...existing, ...user })
        } else {
            this.state.users.set(userId, user)
        }
        return userId
    }

    /**
     * 获取用户信息
     */
    getUser(userId) {
        return this.state.users.get(userId) || null
    }

    /**
     * 批量添加用户
     */
    addUsers(users) {
        users.forEach(user => this.upsertUser(user))
    }

    // ============ 未读消息管理 ============

    /**
     * 增加未读计数
     */
    incrementUnreadCount(sessionId) {
        const current = this.state.unreadCounts.get(sessionId) || 0
        this.state.unreadCounts.set(sessionId, current + 1)

        // 触发未读总数更新
        this.updateTotalUnreadCount()
    }

    /**
     * 减少未读计数
     */
    decrementUnreadCount(sessionId) {
        const current = this.state.unreadCounts.get(sessionId) || 0
        if (current > 0) {
            this.state.unreadCounts.set(sessionId, current - 1)
            this.updateTotalUnreadCount()
        }
    }

    /**
     * 重置未读计数
     */
    resetUnreadCount(sessionId) {
        this.state.unreadCounts.set(sessionId, 0)
        this.updateTotalUnreadCount()
    }

    /**
     * 获取会话未读数
     */
    getUnreadCount(sessionId) {
        return this.state.unreadCounts.get(sessionId) || 0
    }

    /**
     * 获取总未读数
     */
    updateTotalUnreadCount() {
        let total = 0
        for (const count of this.state.unreadCounts.values()) {
            total += count
        }
        this.state.totalUnreadCount = total
    }

    // ============ 搜索功能 ============

    /**
     * 搜索消息
     * @param {string} keyword - 关键词
     * @param {string} sessionId - 限定会话（可选）
     * @param {number} limit - 限制数量
     */
    searchMessages(keyword, sessionId = null, limit = 100) {
        if (!keyword.trim()) return []

        const results = []
        const kw = keyword.toLowerCase()

        // 遍历所有消息（注意：大数据量时可能需要优化）
        for (const [msgId, message] of this.state.messages) {
            // 如果指定了会话，检查是否匹配
            if (sessionId && message.sessionId !== sessionId) continue

            // 搜索内容（可以根据需要扩展搜索字段）
            if (
                (message.content && message.content.toLowerCase().includes(kw)) ||
                (message.senderName && message.senderName.toLowerCase().includes(kw))
            ) {
                results.push(message)

                if (results.length >= limit) break
            }
        }

        // 按时间倒序排序
        return results.sort((a, b) => b.timestamp - a.timestamp)
    }

    // ============ 数据统计 ============

    /**
     * 获取数据统计
     */
    getStatistics() {
        return {
            sessions: this.state.sessions.size,
            messages: this.state.messages.size,
            users: this.state.users.size,
            totalUnread: this.state.totalUnreadCount || 0,
            memoryUsage: this.estimateMemoryUsage()
        }
    }

    /**
     * 估算内存使用
     */
    estimateMemoryUsage() {
        // 简单估算
        let bytes = 0

        // 消息大小
        for (const msg of this.state.messages.values()) {
            bytes += JSON.stringify(msg).length * 2 // UTF-16
        }

        // 会话大小
        for (const session of this.state.sessions.values()) {
            bytes += JSON.stringify(session).length * 2
        }

        // 用户大小
        for (const user of this.state.users.values()) {
            bytes += JSON.stringify(user).length * 2
        }

        return {
            bytes,
            kb: (bytes / 1024).toFixed(2),
            mb: (bytes / (1024 * 1024)).toFixed(2)
        }
    }

    // ============ 数据导入/导出 ============

    /**
     * 导出数据
     */
    exportData() {
        return {
            sessions: Array.from(this.state.sessions.values()),
            messages: Array.from(this.state.messages.values()),
            users: Array.from(this.state.users.values()),
            version: '1.0',
            exportedAt: Date.now()
        }
    }

    /**
     * 导入数据
     */
    importData(data) {
        // 清空现有数据
        this.clear()

        // 导入新数据
        if (data.sessions) {
            data.sessions.forEach(session => this.upsertSession(session))
        }

        if (data.messages) {
            this.addMessages(data.messages)
        }

        if (data.users) {
            this.addUsers(data.users)
        }

        return this.getStatistics()
    }

    /**
     * 清空所有数据
     */
    clear() {
        this.state.sessions.clear()
        this.state.messages.clear()
        this.state.sessionMessages.clear()
        this.state.users.clear()
        this.state.unreadCounts.clear()
        this.state.sortedSessionIds = []
        this.state.activeSessionId = null
        this.state.totalUnreadCount = 0
    }

    // ============ 工具方法 ============

    /**
     * 设置当前用户ID
     */
    setCurrentUserId(userId) {
        this.currentUserId = userId
    }
}

// 创建单例
let messageStoreInstance = null

export function createMessageStore() {
    if (!messageStoreInstance) {
        messageStoreInstance = new MessageStore()
    }
    return messageStoreInstance
}

export function useMessageStore() {
    const store = createMessageStore()

    // 返回响应式状态和方法
    return {
        // 状态
        sessions: computed(() => store.getAllSessions()),
        activeSessionId: computed(() => store.state.activeSessionId),
        activeSession: computed(() =>
            store.state.activeSessionId ? store.getSession(store.state.activeSessionId) : null
        ),
        totalUnreadCount: computed(() => store.state.totalUnreadCount || 0),
        statistics: computed(() => store.getStatistics()),

        // 方法
        upsertSession: store.upsertSession.bind(store),
        getSession: store.getSession.bind(store),
        getAllSessions: store.getAllSessions.bind(store),

        addMessage: store.addMessage.bind(store),
        addMessages: store.addMessages.bind(store),
        getSessionMsg: store.getSessionMsg.bind(store),
        getSessionMessages: store.getSessionMessages.bind(store),
        updateMessage: store.updateMessage.bind(store),
        fetchAllMessage: store.fetchAllMessage.bind(store),

        upsertUser: store.upsertUser.bind(store),
        getUser: store.getUser.bind(store),

        setActiveSession: (sessionId) => {
            store.state.activeSessionId = sessionId
            if (sessionId) {
                store.resetUnreadCount(sessionId)
            }
        },

        getUnreadCount: store.getUnreadCount.bind(store),
        resetUnreadCount: store.resetUnreadCount.bind(store),

        searchMessages: store.searchMessages.bind(store),

        exportData: store.exportData.bind(store),
        importData: store.importData.bind(store),
        clear: store.clear.bind(store),

        setCurrentUserId: store.setCurrentUserId.bind(store)
    }
}