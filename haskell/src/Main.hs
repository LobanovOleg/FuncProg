{-# LANGUAGE OverloadedStrings #-}
{-# LANGUAGE RecordWildCards #-}

import qualified Data.Map as Map
import Text.Megaparsec
import Text.Megaparsec.Char (space1, alphaNumChar, letterChar, string)
import qualified Text.Megaparsec.Char.Lexer as L
import Control.Monad (void, guard)
import Control.Monad.Combinators.Expr
import Data.Void
import System.Environment (getArgs)

type Parser = Parsec Void String

-- AST
data Expr = Number Int
          | Var String
          | BinaryOp String Expr Expr
          | IfThenElse Expr Expr Expr
          | LetRec String [String] Expr Expr
          | Call String [Expr]
  deriving (Show)

-- Лексические помощники
sc :: Parser ()
sc = L.space space1 empty empty

lexeme :: Parser a -> Parser a
lexeme = L.lexeme sc

symbol :: String -> Parser String
symbol s = lexeme (string s)

-- Парсеры
integer :: Parser Int
integer = lexeme L.decimal

parens :: Parser a -> Parser a
parens = between (symbol "(") (symbol ")")

identifier :: Parser String
identifier = lexeme ((:) <$> letterChar <*> many alphaNumChar)

reservedKeywords = ["let", "rec", "if", "then", "else", "in"]

isReserved :: String -> Bool
isReserved = (`elem` reservedKeywords)

-- Базовые выражения
atom :: Parser Expr
atom =  try var
    <|> try number
    <|> try (parens expr)
    <|> try ifThenElse
    <|> try letRec

var :: Parser Expr
var = do
  x <- identifier
  guard (not $ isReserved x)
  return $ Var x

number :: Parser Expr
number = Number <$> integer

ifThenElse :: Parser Expr
ifThenElse = do
  void $ symbol "if"
  cond <- expr
  void $ symbol "then"
  e1 <- expr
  void $ symbol "else"
  e2 <- expr
  return $ IfThenElse cond e1 e2

letRec :: Parser Expr
letRec = do
  void $ symbol "let"
  void $ symbol "rec"
  fname <- identifier
  guard (not $ isReserved fname)
  args <- many $ lexeme $ try $ do
            x <- identifier
            guard (not $ isReserved x)
            return x
  void $ symbol "="
  body <- expr
  void $ symbol "in"
  rest <- expr
  return $ LetRec fname args body rest

-- Вызовы функций
application :: Parser Expr
application = do
  f <- identifier
  guard (not $ isReserved f)
  args <- some atom
  return $ Call f args

factor :: Parser Expr
factor = try application <|> atom

expr :: Parser Expr
expr = makeExprParser factor opTable <?> "expression"

opTable :: [[Operator Parser Expr]]
opTable = 
  [ [ InfixL (BinaryOp "*" <$ symbol "*")
    , InfixL (BinaryOp "/" <$ symbol "/")
    , InfixL (BinaryOp "%" <$ symbol "%")
    ]
  , [ InfixL (BinaryOp "+" <$ symbol "+")
    , InfixL (BinaryOp "-" <$ symbol "-")
    ]
  , [ InfixL (BinaryOp "<=" <$ symbol "<=")
    , InfixL (BinaryOp "<" <$ symbol "<")
    , InfixL (BinaryOp ">=" <$ symbol ">=")
    , InfixL (BinaryOp ">" <$ symbol ">")
    , InfixL (BinaryOp "==" <$ symbol "==")
    , InfixL (BinaryOp "/=" <$ symbol "/=")
    ]
  ]

parseProgram :: String -> Either String Expr
parseProgram input = 
  case parse (sc *> expr <* eof) "(input)" input of
    Left err -> Left $ errorBundlePretty err
    Right e -> Right e

-- Интерпретатор
type Env = Map.Map String Int

data Closure = Closure 
  { closureParams :: [String]
  , closureBody   :: Expr
  , closureEnv    :: Env
  }

type FuncEnv = Map.Map String Closure

eval :: Env -> FuncEnv -> Expr -> Int
eval env _ (Number n) = n
eval env _ (Var x) = case Map.lookup x env of
                       Just v -> v
                       Nothing -> error $ "Variable not found: " ++ x
eval env fenv (BinaryOp op e1 e2) =
  let v1 = eval env fenv e1
      v2 = eval env fenv e2
  in case op of
       "+" -> v1 + v2
       "-" -> v1 - v2
       "*" -> v1 * v2
       "/" -> if v2 == 0 then error "Division by zero" else v1 `div` v2
       "%" -> if v2 == 0 then error "Modulo by zero" else v1 `mod` v2
       "<=" -> if v1 <= v2 then 1 else 0
       "<"  -> if v1 < v2 then 1 else 0
       ">=" -> if v1 >= v2 then 1 else 0
       ">"  -> if v1 > v2 then 1 else 0
       "==" -> if v1 == v2 then 1 else 0
       "/=" -> if v1 /= v2 then 1 else 0
       _ -> error ("Unknown operator: " ++ op)
eval env fenv (IfThenElse cond e1 e2) =
  if eval env fenv cond /= 0 then eval env fenv e1 else eval env fenv e2
eval env fenv (LetRec fname args body rest) =
  let fenv' = Map.insert fname (Closure args body env) fenv
  in eval env fenv' rest
eval env fenv (Call fname argsExprs) =
  case Map.lookup fname fenv of
    Just closure ->
      let Closure {..} = closure
          argVals = map (eval env fenv) argsExprs
          newEnv = foldl (\e (p, v) -> Map.insert p v e) closureEnv (zip closureParams argVals)
      in eval newEnv fenv closureBody
    Nothing -> error $ "Undefined function: " ++ fname

main :: IO ()
main = do
  args <- getArgs
  let fileName = if null args then "examples/basic.fac" else head args
  content <- readFile fileName
  
  case parseProgram content of
    Left err -> putStrLn $ "Parse error: " ++ err
    Right ast -> 
      let result = eval Map.empty Map.empty ast
      in print result