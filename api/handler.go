package api

import (
	"fmt"
	"net/http"

	"github.com/exppii/llmchain/api/log"
	"github.com/exppii/llmchain/api/model"
	"github.com/gin-gonic/gin"
)

func chatEndpointHandler(manager *model.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {

		input, err := parseReq(c)
		if err != nil {
			log.E("failed reading parameters from request: ", err.Error())
			//todo 从中间件拿取语言类型
			c.JSON(http.StatusBadRequest, ReqArgsErr.WithMessage(err.Error()))
			return
		}

		llm, err := manager.GetModel(input.Model)

		if err != nil {

			log.E("model not found or loaded")

			c.JSON(http.StatusInternalServerError, ModelNotExistsErr.WithMessage(err.Error()))
			return
		}

		if input.Stream {
			log.D("Stream request received")

			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")
			c.Header("Transfer-Encoding", "chunked")
		}

		println(llm)

	}
}

func editEndpointHandler(manager *model.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func completionEndpointHandler(manager *model.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {

		log.I(`parse input...`)
		input, err := parseReq(c)
		if err != nil {
			log.E("failed reading parameters from request: ", err.Error())
			//todo 从中间件拿取语言类型
			c.JSON(http.StatusBadRequest, ReqArgsErr.WithMessage(err.Error()))
			return
		}

		llm, err := manager.GetModel(input.Model)

		if err != nil {

			log.E("model not found or loaded")

			c.JSON(http.StatusInternalServerError, ModelNotExistsErr.WithMessage(err.Error()))
			return
		}
		log.I(`merge input to predict options`)
		payload := llm.MergeModelOptions(input)

		templatedInputs, err := payload.TemplatePromptStrings(manager.GetPrompt())

		if err != nil {

			log.E("err when template input data: ", err.Error())
			//TODO
			c.JSON(http.StatusInternalServerError, ModelNotExistsErr.WithMessage(err.Error()))
			return
		}

		fn := func(s string, c *[]Choice) { *c = append(*c, Choice{Text: s}) }

		var result []Choice
		for _, i := range templatedInputs {

			r, err := ComputeChoices(llm, i, payload, fn, nil)
			if err != nil {

				log.E(`error when compute choices: `, err.Error())
				//TODO
				c.JSON(http.StatusInternalServerError, ModelNotExistsErr.WithMessage(err.Error()))
				return
			}

			result = append(result, r...)
		}

		resp := &OpenAIResponse{
			Model:   input.Model, // we have to return what the user sent here, due to OpenAI spec.
			Choices: result,
			Object:  "text_completion",
		}

		// jsonResult, _ := json.Marshal(resp)
		// log.Debug().Msgf("Response: %s", jsonResult)

		// Return the prediction in the response body
		c.JSON(http.StatusOK, resp)

	}
}

func embeddingsEndpointHandler(manager *model.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func listModelsHandler(manager *model.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {

		log.I(`received list model req`)
		list := manager.ListModels()

		models := []OpenAIModel{}
		for _, m := range list {
			models = append(models, OpenAIModel{ID: m, Object: "model"})
		}

		resp := struct {
			Object string        `json:"object"`
			Data   []OpenAIModel `json:"data"`
		}{
			Object: "list",
			Data:   models,
		}

		c.JSON(http.StatusOK, resp)

	}
}

// parseReq parse openai format req
func parseReq(c *gin.Context) (*OpenAIRequest, error) {

	input := &OpenAIRequest{}

	if err := c.Bind(input); err != nil {
		return nil, fmt.Errorf("failed reading parameters from request: ", err.Error())

	}
	return input, nil
}
