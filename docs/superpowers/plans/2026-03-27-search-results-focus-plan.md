# Search Results Focus Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `/search` prioritize the focused result and improve result-page hierarchy without changing backend APIs.

**Architecture:** Keep the existing `/search` route and data shape, then add a small ranking/selection layer in `global-search.js` so `SearchView.vue` can render a best-match card, choose a better default tab, and trim the all-results layout. UI remains a single view with incremental cards and grouped previews.

**Tech Stack:** Vue 3, Vue Router, Vite, Node test runner

---
