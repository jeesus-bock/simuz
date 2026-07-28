# Lua Scripting Guide

This document describes the Lua scripting system used for entity AI behavior in the simulation engine.

## Overview

The AI system allows entities to run Lua scripts that control their behavior each tick. Scripts are loaded at runtime and can interact with the entity, world, and other systems through a set of exposed bindings.

## Script Management

### Loading Scripts

