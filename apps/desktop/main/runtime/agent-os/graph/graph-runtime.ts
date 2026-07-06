import type { GraphEdge } from './graph-edge'
import type { GraphNode } from './graph-node'

export class GraphRuntime {
  private readonly nodes = new Map<string, GraphNode>()
  private readonly edges = new Map<string, GraphEdge>()

  addNode(node: GraphNode): void {
    this.nodes.set(node.id, node)
  }

  addEdge(edge: GraphEdge): void {
    this.edges.set(edge.id, edge)
  }

  snapshot(): { nodes: GraphNode[]; edges: GraphEdge[] } {
    return {
      nodes: [...this.nodes.values()],
      edges: [...this.edges.values()]
    }
  }
}
