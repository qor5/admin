package tag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExpressionSimplify verifies the flattening behavior of nested expressions
// to ensure complex hierarchical expressions are properly normalized
func TestExpressionSimplify(t *testing.T) {
	t.Run("NestedIntersect", func(t *testing.T) {
		nestedIntersect := &Expression{
			Intersect: []*Expression{
				{
					Tag: &Tag{
						BuilderID: "tag1",
						Params: map[string]any{
							"param1": "value1",
						},
					},
				},
				{
					Intersect: []*Expression{
						{
							Tag: &Tag{
								BuilderID: "tag2",
								Params: map[string]any{
									"param2": "value2",
								},
							},
						},
						{
							Tag: &Tag{
								BuilderID: "tag3",
								Params: map[string]any{
									"param3": "value3",
								},
							},
						},
					},
				},
			},
		}

		simplified := nestedIntersect.Simplify()
		assert.Equal(t, 3, len(simplified.Intersect), "Should flatten INTERSECT expressions")
		assert.Equal(t, "tag1", simplified.Intersect[0].Tag.BuilderID)
		assert.Equal(t, "tag2", simplified.Intersect[1].Tag.BuilderID)
		assert.Equal(t, "tag3", simplified.Intersect[2].Tag.BuilderID)
	})

	t.Run("NestedUnion", func(t *testing.T) {
		nestedUnion := &Expression{
			Union: []*Expression{
				{
					Tag: &Tag{
						BuilderID: "tag1",
						Params: map[string]any{
							"param1": "value1",
						},
					},
				},
				{
					Union: []*Expression{
						{
							Tag: &Tag{
								BuilderID: "tag2",
								Params: map[string]any{
									"param2": "value2",
								},
							},
						},
						{
							Tag: &Tag{
								BuilderID: "tag3",
								Params: map[string]any{
									"param3": "value3",
								},
							},
						},
					},
				},
			},
		}

		simplifiedUnion := nestedUnion.Simplify()
		assert.Equal(t, 3, len(simplifiedUnion.Union), "Should flatten UNION expressions")
		assert.Equal(t, "tag1", simplifiedUnion.Union[0].Tag.BuilderID)
		assert.Equal(t, "tag2", simplifiedUnion.Union[1].Tag.BuilderID)
		assert.Equal(t, "tag3", simplifiedUnion.Union[2].Tag.BuilderID)
	})

	t.Run("ExceptOperation", func(t *testing.T) {
		exceptExpr := &Expression{
			Except: []*Expression{
				{
					Tag: &Tag{
						BuilderID: "tag1",
					},
				},
				{
					Tag: &Tag{
						BuilderID: "tag2",
					},
				},
			},
		}

		simplified := exceptExpr.Simplify()
		assert.Equal(t, 2, len(simplified.Except), "Should preserve EXCEPT expressions")
		assert.Equal(t, "tag1", simplified.Except[0].Tag.BuilderID)
		assert.Equal(t, "tag2", simplified.Except[1].Tag.BuilderID)
	})

	t.Run("DeeplyNested", func(t *testing.T) {
		deeplyNested := &Expression{
			Intersect: []*Expression{
				{
					Tag: &Tag{
						BuilderID: "tag1",
					},
				},
				{
					Intersect: []*Expression{
						{
							Tag: &Tag{
								BuilderID: "tag2",
							},
						},
						{
							Intersect: []*Expression{
								{
									Tag: &Tag{
										BuilderID: "tag3",
									},
								},
								{
									Tag: &Tag{
										BuilderID: "tag4",
									},
								},
							},
						},
					},
				},
			},
		}

		simplifiedDeep := deeplyNested.Simplify()
		assert.Equal(t, 4, len(simplifiedDeep.Intersect), "Should flatten deeply nested INTERSECT expressions")
		assert.Equal(t, "tag1", simplifiedDeep.Intersect[0].Tag.BuilderID)
		assert.Equal(t, "tag2", simplifiedDeep.Intersect[1].Tag.BuilderID)
		assert.Equal(t, "tag3", simplifiedDeep.Intersect[2].Tag.BuilderID)
		assert.Equal(t, "tag4", simplifiedDeep.Intersect[3].Tag.BuilderID)
	})

	t.Run("MixedTypes", func(t *testing.T) {
		mixedTypes := &Expression{
			Intersect: []*Expression{
				{
					Tag: &Tag{
						BuilderID: "tag1",
					},
				},
				{
					Union: []*Expression{
						{
							Tag: &Tag{
								BuilderID: "tag2",
							},
						},
						{
							Tag: &Tag{
								BuilderID: "tag3",
							},
						},
					},
				},
			},
		}

		simplifiedMixed := mixedTypes.Simplify()
		assert.Equal(t, 2, len(simplifiedMixed.Intersect), "Should preserve different expression types")
		assert.Equal(t, "tag1", simplifiedMixed.Intersect[0].Tag.BuilderID)
		assert.Equal(t, 2, len(simplifiedMixed.Intersect[1].Union), "Should preserve UNION in INTERSECT")
	})

	t.Run("Advanced", func(t *testing.T) {
		advanced := &Expression{
			Intersect: []*Expression{
				{
					Tag: &Tag{
						BuilderID: "tag1",
					},
				},
				{
					Intersect: []*Expression{
						{
							Intersect: []*Expression{
								{
									Tag: &Tag{
										BuilderID: "tag2",
									},
								},
								{
									Tag: &Tag{
										BuilderID: "placeholder",
									},
								},
							},
						},
						{
							Tag: &Tag{
								BuilderID: "tag3",
							},
						},
						{
							Intersect: []*Expression{
								{
									Tag: &Tag{
										BuilderID: "tag4",
									},
								},
								{
									Union: []*Expression{
										{
											Tag: &Tag{
												BuilderID: "tag5",
											},
										},
										{
											Tag: &Tag{
												BuilderID: "tag6",
											},
										},
									},
								},
								{
									Except: []*Expression{
										{
											Intersect: []*Expression{
												{
													Intersect: []*Expression{
														{
															Tag: &Tag{
																BuilderID: "tag7",
															},
														},
														{
															Tag: &Tag{
																BuilderID: "tag8",
															},
														},
													},
												},
											},
										},
										{
											Intersect: []*Expression{
												{
													Tag: &Tag{
														BuilderID: "tag9",
													},
												},
											},
										},
										{
											Except: []*Expression{
												{
													Tag: &Tag{
														BuilderID: "tag10",
													},
												},
												{
													Tag: &Tag{
														BuilderID: "tag11",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		simplifiedAdvanced := advanced.Simplify()

		// Check if nested INTERSECT expressions are flattened correctly
		assert.Equal(t, 7, len(simplifiedAdvanced.Intersect), "Should correctly flatten nested INTERSECT expressions")

		// Check first five elements are individual tags
		assert.Equal(t, "tag1", simplifiedAdvanced.Intersect[0].Tag.BuilderID)
		assert.Equal(t, "tag2", simplifiedAdvanced.Intersect[1].Tag.BuilderID)
		assert.Equal(t, "placeholder", simplifiedAdvanced.Intersect[2].Tag.BuilderID)
		assert.Equal(t, "tag3", simplifiedAdvanced.Intersect[3].Tag.BuilderID)
		assert.Equal(t, "tag4", simplifiedAdvanced.Intersect[4].Tag.BuilderID)

		// Check the sixth element is a UNION expression
		assert.Equal(t, 2, len(simplifiedAdvanced.Intersect[5].Union), "Should preserve UNION expression")
		assert.Equal(t, "tag5", simplifiedAdvanced.Intersect[5].Union[0].Tag.BuilderID)
		assert.Equal(t, "tag6", simplifiedAdvanced.Intersect[5].Union[1].Tag.BuilderID)

		// Check the seventh element is an EXCEPT expression
		assert.Equal(t, 3, len(simplifiedAdvanced.Intersect[6].Except), "Should preserve EXCEPT expression")

		// First element in Except should be an Intersect with tag7 and tag8
		assert.Equal(t, 2, len(simplifiedAdvanced.Intersect[6].Except[0].Intersect))
		assert.Equal(t, "tag7", simplifiedAdvanced.Intersect[6].Except[0].Intersect[0].Tag.BuilderID)
		assert.Equal(t, "tag8", simplifiedAdvanced.Intersect[6].Except[0].Intersect[1].Tag.BuilderID)

		// Second element in Except should be a Tag with BuilderID tag8
		assert.Equal(t, "tag9", simplifiedAdvanced.Intersect[6].Except[1].Tag.BuilderID)

		// Third element in Except should be an EXCEPT expression with tag10 and tag11
		assert.Equal(t, 2, len(simplifiedAdvanced.Intersect[6].Except[2].Except))
		assert.Equal(t, "tag10", simplifiedAdvanced.Intersect[6].Except[2].Except[0].Tag.BuilderID)
		assert.Equal(t, "tag11", simplifiedAdvanced.Intersect[6].Except[2].Except[1].Tag.BuilderID)
	})
}
